package pegoutmanager

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	entpegout "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
)

const (
	RBF_SEQUENCE                       uint32 = 0xFFFFFFFD
	defaultLoopInterval                       = 3 * time.Second
	semaphoreLimit                            = 32
	pegoutQueryLimit                          = 1024
	twoWeeks                                  = 14 * 24 * time.Hour
	badTxnsInputsMissingOrSpentErr            = "bad-txns-inputs-missingorspent"
	unspendableOutputExceedsMaxBurnErr        = "unspendable output exceeds maximum configured by user"
)

type PegoutManager struct {
	ctx                        context.Context
	repo                       *ent.Client
	bitcoinClient              *bitcoin.Client
	tonClient                  *tonclient.TonClient
	teleportContract           *teleportcontract.TeleportContract
	excludedSignedPegoutIDs    map[int]struct{}
	excludedSignedPegoutIDsMux sync.Mutex
}

func New(
	ctx context.Context,
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	teleportContract *teleportcontract.TeleportContract,
) (
	*PegoutManager,
	error,
) {
	pegoutManager := &PegoutManager{
		ctx:                     ctx,
		repo:                    repo,
		bitcoinClient:           bitcoinClient,
		tonClient:               tonClient,
		teleportContract:        teleportContract,
		excludedSignedPegoutIDs: make(map[int]struct{}),
	}

	return pegoutManager, nil
}

func (pm *PegoutManager) Run() error {
	pm.logStartManager()

	var wg sync.WaitGroup
	wg.Add(2)

	go pm.continuouslyProcessSigningPegouts(&wg)
	go pm.continuouslyProcessSignedPegouts(&wg)

	<-pm.ctx.Done()
	logContextCancelled()

	wg.Wait()
	pm.logStopManager(pm.ctx.Err())
	return pm.ctx.Err()
}

func (pm *PegoutManager) continuouslyProcessSigningPegouts(wg *sync.WaitGroup) {
	defer wg.Done()
	pm.logStartSigningWork()
	var err error
	defer func() { pm.logFinishSigningWork(err) }()

	ticker := time.NewTicker(defaultLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			err = pm.ctx.Err()
			return
		case <-ticker.C:
			cycleErr := pm.executeSigningPegoutsCycle()
			if cycleErr != nil {
				logSigningCycleError(cycleErr)
			}
		}
	}
}

func (pm *PegoutManager) executeSigningPegoutsCycle() (err error) {
	start := time.Now()
	var processedCount int
	defer func() {
		logFinishProcessingSigningPegouts(time.Since(start), err, processedCount)
	}()
	logStartProcessingSigningPegouts()

	pegouts, err := pm.repo.Pegout.Query().
		Where(
			entpegout.StatusEQ(entpegout.StatusSigning),
			entpegout.AddrNotNil(),
			entpegout.AddrNotIn("NONE", "NONE1")).
		Limit(pegoutQueryLimit).
		All(pm.ctx)
	if err != nil {
		return fmt.Errorf(errQuerySigningPegouts, err)
	}

	if len(pegouts) == 0 {
		logNoSigningPegouts()
		return nil
	}
	logSigningPegoutsReceived(len(pegouts))
	processedCount = len(pegouts)

	block, err := pm.tonClient.API.CurrentMasterchainInfo(pm.ctx)
	if err != nil {
		return fmt.Errorf("failed to get current masterchain info for signing cycle: %w", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, semaphoreLimit)

	for i, pegout := range pegouts {
		wg.Add(1)
		sem <- struct{}{}
		go func(p *ent.Pegout, idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := pm.handleSigningPegout(block, p); err != nil {
				logFailedProcessSigningPegout(err, p.Addr)
			}
			logSigningPegoutProgress(idx+1, len(pegouts))
		}(pegout, i)
	}

	wg.Wait()
	return nil
}

func (pm *PegoutManager) continuouslyProcessSignedPegouts(wg *sync.WaitGroup) {
	defer wg.Done()
	pm.logStartSignedWork()
	var err error
	defer func() { pm.logFinishSignedWork(err) }()

	ticker := time.NewTicker(defaultLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			err = pm.ctx.Err()
			return
		case <-ticker.C:
			cycleErr := pm.executeSignedPegoutsCycle()
			if cycleErr != nil {
				logSignedCycleError(cycleErr)
			}
		}
	}
}

func (pm *PegoutManager) executeSignedPegoutsCycle() (err error) {
	start := time.Now()
	var processedInCycleCount int
	var excludedThisCycleCount int
	var totalInMemoryExclusions int

	defer func() {
		pm.excludedSignedPegoutIDsMux.Lock()
		totalInMemoryExclusions = len(pm.excludedSignedPegoutIDs)
		pm.excludedSignedPegoutIDsMux.Unlock()
		logFinishProcessingSignedPegouts(time.Since(start), err, processedInCycleCount, excludedThisCycleCount, totalInMemoryExclusions)
	}()
	logStartProcessingSignedPegouts()

	pm.excludedSignedPegoutIDsMux.Lock()
	excludedIDs := make([]int, 0, len(pm.excludedSignedPegoutIDs))
	for id := range pm.excludedSignedPegoutIDs {
		excludedIDs = append(excludedIDs, id)
	}
	totalInMemoryExclusions = len(pm.excludedSignedPegoutIDs)
	pm.excludedSignedPegoutIDsMux.Unlock()

	query := pm.repo.Pegout.Query().
		Where(
			entpegout.StatusEQ(entpegout.StatusSigned),
			entpegout.AddrNotNil(),
			entpegout.AddrNotIn("NONE", "NONE1")).
		Limit(pegoutQueryLimit)

	if len(excludedIDs) > 0 {
		query = query.Where(entpegout.IDNotIn(excludedIDs...))
	}

	allPegouts, err := query.All(pm.ctx)
	if err != nil {
		return fmt.Errorf(errQuerySignedPegouts, err)
	}

	if len(allPegouts) == 0 {
		logNoSignedPegouts()
		return nil
	}

	processList := make([]*ent.Pegout, 0, len(allPegouts))
	pm.excludedSignedPegoutIDsMux.Lock()
	for _, p := range allPegouts {
		if _, excluded := pm.excludedSignedPegoutIDs[p.ID]; excluded {
			excludedThisCycleCount++
			continue
		}
		processList = append(processList, p)
	}
	pm.excludedSignedPegoutIDsMux.Unlock()

	if len(processList) == 0 {
		logSignedPegoutsReceived(0)
		return nil
	}
	logSignedPegoutsReceived(len(processList))
	processedInCycleCount = len(processList)

	var wg sync.WaitGroup
	sem := make(chan struct{}, semaphoreLimit)

	for i, pegout := range processList {
		wg.Add(1)
		sem <- struct{}{}
		go func(p *ent.Pegout, idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := pm.handleSignedPegout(p); err != nil {
				logFailedProcessSignedPegout(err, p.Addr, p.BitcoinTxID)
			}
			logSignedPegoutProgress(idx+1, len(processList))
		}(pegout, i)
	}

	wg.Wait()
	return nil
}

func (pm *PegoutManager) handleSigningPegout(
	block *ton.BlockIDExt,
	pegout *ent.Pegout,
) error {
	pegoutState, err := pm.tonClient.API.GetAccount(pm.ctx, block, address.MustParseRawAddr(pegout.Addr))
	if err != nil {
		return fmt.Errorf(errGetPegoutState, err)
	}
	if !pegoutState.IsActive {
		logPegoutContractNotActive(pegout.Addr)
		return nil
	}

	pegoutContract := pegoutcontract.New(
		address.MustParseRawAddr(pegout.Addr),
		pm.tonClient,
		pm.ctx,
	)

	pegoutTxParts, err := pegoutContract.GetTxParts(block)
	if err != nil {
		return fmt.Errorf(errGetPegoutTxParts, err)
	}

	pegoutTx, err := pm.buildPegoutTx(pegoutTxParts)
	if err != nil {
		return fmt.Errorf(errBuildPegoutTx, err)
	}

	txHex, err := bitcoin.TxToHex(pegoutTx)
	if err != nil {
		return fmt.Errorf(errSerializePegoutTx, err)
	}

	updateErr := pm.repo.Pegout.Update().
		SetStatus(entpegout.StatusSigned).
		SetBitcoinTxRaw(txHex).
		SetBitcoinTxID(pegoutTx.TxID()).
		Where(entpegout.ID(pegout.ID)).
		Exec(pm.ctx)
	if updateErr != nil {
		logFailedUpdatePegoutStatus(updateErr, pegout.ID, entpegout.StatusSigned, pegoutTx.TxID())
		return fmt.Errorf(errUpdatePegoutToSigned, updateErr)
	}
	logPegoutStatusUpdated(pegout.ID, entpegout.StatusSigned, pegoutTx.TxID())
	return nil
}

func (pm *PegoutManager) handleSignedPegout(
	pegout *ent.Pegout,
) error {
	txHash, err := chainhash.NewHashFromStr(pegout.BitcoinTxID)
	if err != nil {
		return fmt.Errorf(errParseTxHash, err)
	}
	txVerbose, err := pm.bitcoinClient.GetRawTransactionVerbose(txHash)
	if err == nil {
		updateErr := pm.repo.Pegout.Update().
			SetStatus(entpegout.StatusConfirmed).
			SetBitcoinBlockHash(txVerbose.BlockHash).
			Where(entpegout.ID(pegout.ID)).
			Exec(pm.ctx)
		if updateErr != nil {
			logFailedUpdatePegoutStatus(updateErr, pegout.ID, entpegout.StatusConfirmed, pegout.BitcoinTxID)
			return fmt.Errorf(errUpdatePegoutToConfirm, updateErr)
		}
		logPegoutStatusUpdated(pegout.ID, entpegout.StatusConfirmed, pegout.BitcoinTxID)
		return nil
	}

	tx, err := bitcoin.HexToTx(pegout.BitcoinTxRaw, 2)
	if err != nil {
		return fmt.Errorf(errDeserializePegoutTx, err)
	}
	_, sendErr := pm.bitcoinClient.SendRawTransaction(tx, false)
	if sendErr != nil {
		sendErrStr := strings.ToLower(sendErr.Error())
		if strings.Contains(sendErrStr, badTxnsInputsMissingOrSpentErr) ||
			strings.Contains(sendErrStr, unspendableOutputExceedsMaxBurnErr) {
			pWithEdges, fetchErr := pm.repo.Pegout.Query().
				Where(entpegout.ID(pegout.ID)).
				WithBurn(func(bq *ent.BurnQuery) { bq.WithTonTx() }).
				WithReinit(func(rq *ent.ReinitQuery) { rq.WithTonTx() }).
				Only(pm.ctx)

			if fetchErr != nil {
				logCouldNotDeterminePegoutAge(pegout.ID, pegout.BitcoinTxID, fetchErr)
			} else {
				var relevantTime time.Time
				foundTime := false
				if pWithEdges.Edges.Burn != nil && pWithEdges.Edges.Burn.Edges.TonTx != nil {
					relevantTime = pWithEdges.Edges.Burn.Edges.TonTx.CreatedAt
					foundTime = true
				} else if pWithEdges.Edges.Reinit != nil && pWithEdges.Edges.Reinit.Edges.TonTx != nil {
					relevantTime = pWithEdges.Edges.Reinit.Edges.TonTx.CreatedAt
					foundTime = true
				}

				if foundTime {
					if time.Since(relevantTime) > twoWeeks {
						pm.excludedSignedPegoutIDsMux.Lock()
						if _, alreadyExcluded := pm.excludedSignedPegoutIDs[pegout.ID]; !alreadyExcluded {
							pm.excludedSignedPegoutIDs[pegout.ID] = struct{}{}
							logSignedPegoutExcluded(pegout.ID, pegout.BitcoinTxID, sendErr.Error())
						}
						pm.excludedSignedPegoutIDsMux.Unlock()
					}
				} else {
					logCouldNotDeterminePegoutAge(pegout.ID, pegout.BitcoinTxID, nil)
				}
			}
		}
		return fmt.Errorf(errSendPegoutTx, sendErr)
	}
	logPegoutTxSent(pegout.Addr, pegout.BitcoinTxID)
	return nil
}

func (pm *PegoutManager) buildPegoutTx(txParts *pegoutcontract.TxParts) (*wire.MsgTx, error) {
	packet, err := psbt.NewFromUnsignedTx(wire.NewMsgTx(2))
	if err != nil {
		return nil, err
	}

	for _, output := range txParts.Outputs {
		txOut := wire.NewTxOut(int64(output.Amount), output.BitcoinScript)
		packet.UnsignedTx.AddTxOut(txOut)
		packet.Outputs = append(packet.Outputs, psbt.POutput{})
	}

	inputTxIDs := make([]string, 0, len(*txParts.Inputs))
	for txid := range *txParts.Inputs {
		inputTxIDs = append(inputTxIDs, txid)
	}
	sort.Slice(inputTxIDs, func(i, j int) bool {
		txidIBytes, errI := hex.DecodeString(inputTxIDs[i])
		txidJBytes, errJ := hex.DecodeString(inputTxIDs[j])
		if errI != nil || errJ != nil {
			return inputTxIDs[i] < inputTxIDs[j]
		}
		return bytes.Compare(txidIBytes, txidJBytes) < 0
	})

	for i, inputTxID := range inputTxIDs {
		input := (*txParts.Inputs)[inputTxID]
		keyBytes, err := hex.DecodeString(inputTxID)
		if err != nil {
			return nil, fmt.Errorf("error decoding input txID for hash: %s, error: %v", inputTxID, err)
		}
		hash, err := chainhash.NewHash(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid previous tx hash from ID %s: %v", inputTxID, err)
		}

		txIn := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: input.Index,
		}, nil, nil)
		txIn.Sequence = RBF_SEQUENCE
		packet.UnsignedTx.AddTxIn(txIn)

		pInput := psbt.PInput{
			WitnessUtxo: &wire.TxOut{
				Value:    input.Amount.Int64(),
				PkScript: input.BitcoinScript,
			},
			TaprootInternalKey: txParts.InternalKey,
		}
		if len(input.BitcoinMerkleRoot) > 0 {
			pInput.TaprootMerkleRoot = input.BitcoinMerkleRoot
		}
		signature := (*txParts.Signatures)[strconv.Itoa(i)]
		if len(signature) < 64 {
			return nil, fmt.Errorf("signature for input %d is too short (length %d)", i, len(signature))
		}
		pInput.TaprootKeySpendSig = signature[len(signature)-64:]
		packet.Inputs = append(packet.Inputs, pInput)
	}

	if err := psbt.MaybeFinalizeAll(packet); err != nil {
		return nil, fmt.Errorf("failed to finalize inputs: %w", err)
	}

	finalizedTx, err := psbt.Extract(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to extract finalized tx: %w", err)
	}

	return finalizedTx, nil
}
