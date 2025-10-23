package metrics

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type MetricsSystemInfo struct {
	mu                     sync.Mutex
	lastUnsignedPegoutInfo *SysPegoutSigningInfo
}

func (systemInfo *MetricsSystemInfo) SystemInfoJson(
	dataSourceDB *data_sources.DataSourceDB,
	alertManager *alerts.AlertManager,
	contractAddrs map[string]*address.Address,
	bitcoinClient *bitcoin.Client,
) (string, error) {
	sysDkgInfo, dkgStatus, err := systemInfo.SysDkgInfo(dataSourceDB, alertManager, contractAddrs)
	if err != nil {
		return "", err
	}

	sysLastPegoutTxInfo, err := systemInfo.SysLastPegoutTxInfo(dataSourceDB, bitcoinClient)
	if err != nil {
		return "", err
	}

	sysPegoutSigningInfo, err := systemInfo.PegoutSigningInfo(dataSourceDB, alertManager, dkgStatus)
	if err != nil {
		return "", err
	}

	sysBalancesInfo, err := systemInfo.BalancesInfo(dataSourceDB, alertManager, contractAddrs)
	if err != nil {
		return "", err
	}

	sysTeleportInfo, err := systemInfo.TeleportInfo(dataSourceDB)
	if err != nil {
		return "", err
	}

	info := SystemInfo{
		DkgInfo:           sysDkgInfo,
		LastPegoutTxInfo:  sysLastPegoutTxInfo,
		PegoutSigningInfo: sysPegoutSigningInfo,
		BalancesInfo:      sysBalancesInfo,
		TeleportInfo:      sysTeleportInfo,
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (systemInfo *MetricsSystemInfo) SysDkgInfo(
	dataSourceDB *data_sources.DataSourceDB,
	alertManager *alerts.AlertManager,
	contractAddrs map[string]*address.Address,
) (*SysDkgInfo, *DkgStatus, error) {
	dkg, err := dataSourceDB.Dkg()
	if err != nil {
		return nil, nil, err
	}

	// restarts, restartsSeverity
	dkgRestartsAlertState, err := alertManager.GetAlertState("alert_dkg_restarts")
	if err != nil {
		return nil, nil, err
	}

	// culprits, culpritsSeverity
	dkgCulpritsAlertState, err := alertManager.GetAlertState("alert_dkg_culprit_found")
	if err != nil {
		return nil, nil, err
	}

	// DkgStatus
	dkgStatus, err := systemInfo.DkgStatus(dataSourceDB)
	if err != nil {
		return nil, nil, err
	}

	return &SysDkgInfo{
		State:            dkg.State,
		StateName:        dkg.State.String(),
		Restarts:         int(dkgRestartsAlertState.Values["restarts"].(int64)),
		RestartsSeverity: dkgRestartsAlertState.Severity,
		CulpritsIdx:      dkgCulpritsAlertState.Values["culprit_ids"].([]int),
		CulpritsSeverity: dkgCulpritsAlertState.Severity,
		Until:            dkg.Until,

		ValidatorsMax:         dkgStatus.DkgInfo.ValidatorsCountMax,
		ValidatorsActive:      dkgStatus.DkgInfo.ValidatorsCountInDkg,
		ValidatorsActiveIdx:   mutils.ExtractMapKeys(dkgStatus.DkgInfo.ValidatorsIdxInDkg),
		ValidatorsInactive:    dkgStatus.DkgInfo.ValidatorsCountNotInDkg,
		ValidatorsInactiveIdx: mutils.ExtractMapKeys(dkgStatus.DkgInfo.ValidatorsIdxNotInDkg),
		ValidatorsEvicted:     dkgStatus.DkgInfo.ValidatorsCountEvicted,
		ValidatorsEvictedIdx:  mutils.ExtractMapKeys(dkgStatus.DkgInfo.ValidatorsIdxEvicted),

		Timeout: 0, // TODO: implement
	}, dkgStatus, nil
}

func (systemInfo *MetricsSystemInfo) SysLastPegoutTxInfo(
	dataSourceDB *data_sources.DataSourceDB,
	bitcoinClient *bitcoin.Client,
) (*SysLastPegoutTxInfo, error) {
	lastSignedPegout, err := dataSourceDB.LastSignedPegout()
	if err != nil {
		return nil, err
	}

	if lastSignedPegout == nil {
		lastSignedPegout, err = dataSourceDB.LastConfirmedPegout()
		if err != nil {
			return nil, err
		}
	}

	// BtcTxStatus
	btcTxStatus := BTC_TX_NOT_PUBLISHED
	bitcoinMempoolTime := int64(0)
	cpfpLength := 0

	if lastSignedPegout != nil {
		btcTxTimestamp := int64(0)

		{
			txHash := mutils.BytesToBTCHash(lastSignedPegout.BitcoinTxId)
			btcMempoolEntry, err := bitcoinClient.RPCClient.GetMempoolEntry(txHash.String())
			if err == nil {
				if btcMempoolEntry != nil {
					btcTxTimestamp = btcMempoolEntry.Time
					btcTxStatus = BTC_TX_IN_MEMPOOL

					// Detect CPFP length
					cpfpLength, err = mutils.BtcGetCPFPChainSize(bitcoinClient, txHash)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		// Check block
		if btcTxStatus == BTC_TX_NOT_PUBLISHED {
			btcBlockHash, err := bitcoinClient.GetBlockHashByTxID(mutils.BytesToBTCHash(lastSignedPegout.BitcoinTxId))
			if err == nil {
				if btcBlockHash != nil {
					btcTxStatus = BTC_TX_IN_BLOCK
				}
			}
		}

		// BitcoinMempoolTime
		if (btcTxStatus == BTC_TX_NOT_PUBLISHED) || (btcTxStatus == BTC_TX_IN_MEMPOOL) {
			bitcoinMempoolTime = int64(time.Duration(time.Now().Unix()-btcTxTimestamp) * time.Second)
		}
	}

	return &SysLastPegoutTxInfo{
		BtcTxStatus:        btcTxStatus,
		BitcoinMempoolTime: int(bitcoinMempoolTime),
		CPFP:               cpfpLength,
	}, nil
}

func (systemInfo *MetricsSystemInfo) PegoutSigningInfo(
	dataSourceDB *data_sources.DataSourceDB,
	alertManager *alerts.AlertManager,
	dkgStatus *DkgStatus,
) (*SysPegoutSigningInfo, error) {
	coordinatorStorage, err := dataSourceDB.CoordinatorContractStorage()
	if err != nil {
		return nil, err
	}

	prevDkg, err := dataSourceDB.PrevDkg()
	if err != nil {
		return nil, err
	}

	if len(coordinatorStorage.UnsignedPegouts) == 0 {
		systemInfo.mu.Lock()
		defer systemInfo.mu.Unlock()
		if systemInfo.lastUnsignedPegoutInfo != nil {
			systemInfo.lastUnsignedPegoutInfo.IsSigned = SIGNING_STATUS_SIGNED
			return systemInfo.lastUnsignedPegoutInfo, nil
		}

		return &SysPegoutSigningInfo{
			Id:                           0,
			Restarts:                     0,
			RestartsSeverity:             alerts.SEVERITY_OK,
			Until:                        time.Unix(0, 0),
			QueueLength:                  0,
			IsSigned:                     NO_PEGOUT,
			IsAutopegout:                 false,
			Signers:                      -1,
			SignersMax:                   int(prevDkg.R3.Count),
			SignersCommitmentActive:      0,
			SignersCommitmentActiveIdx:   make([]int, 0),
			SignersCommitmentInactive:    0,
			SignersCommitmentInactiveIdx: make([]int, 0),
			SignersEvicted:               0,
			SignersEvictedIdx:            make([]int, 0),
			IsInternalKeyCorrect:         true,
			IsInternalKeyCorrectStr:      "Internal Key Is Correct",
		}, nil
	}

	unsignedPegout := coordinatorStorage.UnsignedPegouts[0]

	// restarts
	pegoutRestartsAlertState, err := alertManager.GetAlertState("alert_pegout_restarts")
	if err != nil {
		return nil, err
	}

	isInternalKeyCorrect := bytes.Equal(unsignedPegout.InternalKey, prevDkg.R3.Data.InternalKey)
	var isInternalKeyCorrectStr string
	if isInternalKeyCorrect {
		isInternalKeyCorrectStr = "Internal Key Is Correct"
	} else {
		isInternalKeyCorrectStr = "Internal Key Is INCORRECT"
	}

	// signers bitmask
	signersCommitmentActiveIdx := make([]int, 0)
	signersCommitmentInactiveIdx := make([]int, 0)
	signersEvictedIdx := make([]int, 0)

	for i := 0; i < dkgStatus.PrevDkgInfo.ValidatorsCountMax; i++ {
		if prevDkg.R3.Mask.Bit(i) == 0 {
			continue
		}

		if unsignedPegout.ClaimsMask.Bit(i) == 1 {
			signersEvictedIdx = append(signersEvictedIdx, i)
			continue
		}

		if unsignedPegout.SigningMask.Bit(i) == 1 {
			signersCommitmentActiveIdx = append(signersCommitmentActiveIdx, i)
		} else {
			signersCommitmentInactiveIdx = append(signersCommitmentInactiveIdx, i)
		}
	}

	systemInfo.mu.Lock()
	defer systemInfo.mu.Unlock()
	systemInfo.lastUnsignedPegoutInfo = &SysPegoutSigningInfo{
		Id:                           int(unsignedPegout.ID),
		Restarts:                     int(pegoutRestartsAlertState.Values["restarts"].(int64)),
		RestartsSeverity:             pegoutRestartsAlertState.Severity,
		Until:                        unsignedPegout.ExpiredAt,
		QueueLength:                  len(coordinatorStorage.UnsignedPegouts),
		IsSigned:                     SIGNING_STATUS_NOT_SIGNED,
		IsAutopegout:                 unsignedPegout.IsAutopegout,
		Signers:                      mutils.Popcnt(unsignedPegout.SigningMask),
		SignersMax:                   int(dkgStatus.PrevDkgInfo.ValidatorsCountInDkg),
		SignersCommitmentActive:      len(signersCommitmentActiveIdx),
		SignersCommitmentActiveIdx:   signersCommitmentActiveIdx,
		SignersCommitmentInactive:    len(signersCommitmentInactiveIdx),
		SignersCommitmentInactiveIdx: signersCommitmentInactiveIdx,
		SignersEvicted:               len(signersEvictedIdx),
		SignersEvictedIdx:            signersEvictedIdx,
		IsInternalKeyCorrect:         isInternalKeyCorrect,
		IsInternalKeyCorrectStr:      isInternalKeyCorrectStr,
	}

	return systemInfo.lastUnsignedPegoutInfo, nil
}

func (systemInfo *MetricsSystemInfo) BalancesInfo(
	dataSourceDB *data_sources.DataSourceDB,
	alertManager *alerts.AlertManager,
	contractAddrs map[string]*address.Address,
) (*SysBalancesInfo, error) {
	coordinator, err := dataSourceDB.ActualContractBalance("coordinator")
	if err != nil {
		return nil, err
	}
	coordinatorAlertState, err := alertManager.GetAlertState("alert_contract_balance_coordinator")
	if err != nil {
		return nil, err
	}

	teleport, err := dataSourceDB.ActualContractBalance("teleport")
	if err != nil {
		return nil, err
	}
	teleportAlertState, err := alertManager.GetAlertState("alert_contract_balance_teleport")
	if err != nil {
		return nil, err
	}

	bitclient, err := dataSourceDB.ActualContractBalance("bitclient")
	if err != nil {
		return nil, err
	}
	bitclientAlertState, err := alertManager.GetAlertState("alert_contract_balance_bitclient")
	if err != nil {
		return nil, err
	}

	minter, err := dataSourceDB.ActualContractBalance("minter")
	if err != nil {
		return nil, err
	}
	minterAlertState, err := alertManager.GetAlertState("alert_contract_balance_minter")
	if err != nil {
		return nil, err
	}

	relayer, err := dataSourceDB.ActualContractBalance("relayer")
	if err != nil {
		return nil, err
	}
	relayerAlertState, err := alertManager.GetAlertState("alert_contract_balance_relayer")
	if err != nil {
		return nil, err
	}

	return &SysBalancesInfo{
		Coordinator:         coordinator,
		CoordinatorStr:      mutils.NanoIntToString(coordinator),
		CoordinatorSeverity: coordinatorAlertState.Severity,
		CoordinatorAddr:     contractAddrs["coordinator"],
		Teleport:            teleport,
		TeleportStr:         mutils.NanoIntToString(teleport),
		TeleportAddr:        contractAddrs["teleport"],
		TeleportSeverity:    teleportAlertState.Severity,
		Bitclient:           bitclient,
		BitclientStr:        mutils.NanoIntToString(bitclient),
		BitclientAddr:       contractAddrs["bitclient"],
		BitclientSeverity:   bitclientAlertState.Severity,
		Minter:              minter,
		MinterStr:           mutils.NanoIntToString(minter),
		MinterAddr:          contractAddrs["minter"],
		MinterSeverity:      minterAlertState.Severity,
		Relayer:             relayer,
		RelayerStr:          mutils.NanoIntToString(relayer),
		RelayerAddr:         contractAddrs["relayer"],
		RelayerSeverity:     relayerAlertState.Severity,
	}, nil
}

func (systemInfo *MetricsSystemInfo) TeleportInfo(
	dataSourceDB *data_sources.DataSourceDB,
) (*SysTeleportInfo, error) {
	teleportStorge, err := dataSourceDB.TeleportContractStorage()
	if err != nil {
		return nil, err
	}

	prevDkg, err := dataSourceDB.PrevDkg()
	if err != nil {
		return nil, err
	}

	// Check internal keys
	teleportInternalKey, err := hex.DecodeString(teleportStorge.InternalKey)
	if err != nil {
		return nil, err
	}

	isInternalKeyCorrect := bytes.Equal(teleportInternalKey, prevDkg.R3.Data.InternalKey)
	var isInternalKeyCorrectStr string
	if isInternalKeyCorrect {
		isInternalKeyCorrectStr = "The same input internal key"
	} else {
		isInternalKeyCorrectStr = "DIFFERENT INTERNAL KEYS"
	}

	return &SysTeleportInfo{
		UTXO:                       len(teleportStorge.UTXOset),
		IsSameInputInternalKey:     isInternalKeyCorrect,
		IsSameInputInternalKeyStr:  isInternalKeyCorrectStr,
		TimeSinceLastAutopegout:    300,     // TODO: implement
		TimeSinceLastAutopegoutStr: "5 min", // TODO: implement
		ServiceFee:                 int(teleportStorge.TotalServiceFee),
		LastConfirmed:              -1, // TODO: implement
		LastBtc_LastTon:            -1, // TODO: implement
		LastTon_PegoutBlock:        -1, // TODO: implement
	}, nil
}

func (systemInfo *MetricsSystemInfo) DkgStatus(
	dataSourceDB *data_sources.DataSourceDB,
) (*DkgStatus, error) {
	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.Dkg() begin")
	dkg, err := dataSourceDB.Dkg()
	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.Dkg() end")
	if err != nil {
		return nil, err
	}

	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.DkgBeforeRestart() begin")
	lastRestartDkg, err := dataSourceDB.DkgBeforeRestart(dkg.Until)
	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.DkgBeforeRestart() end")
	if err != nil {
		return nil, err
	}

	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.PrevDkg() begin")
	prevDkg, err := dataSourceDB.PrevDkg()
	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.PrevDkg() end")
	if err != nil {
		return nil, err
	}

	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.CoordinatorContractStorage() begin")
	coordinatorContractData, err := dataSourceDB.CoordinatorContractStorage()
	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("dataSourceDB.CoordinatorContractStorage() end")
	if err != nil {
		return nil, err
	}

	var status DkgStatus
	status.StandaloneMode = coordinatorContractData.StandaloneMode
	status.Original.Dkg = dkg
	status.Original.LastRestartDkg = lastRestartDkg
	status.Original.PrevDkg = prevDkg

	sumarizeDkgInfo := func(dkg *coordinator.DKG) (*DkgInfo, error) {
		logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("sumarizeDkgInfo begin")
		var info DkgInfo

		info.State = dkg.State.String()
		info.VSetSize = len(dkg.VSet)

		if dkg.R1 != nil {
			info.ValidatorsCountInDkg = int(dkg.R1.Count)
		} else {
			info.ValidatorsCountInDkg = 0
		}

		if !coordinatorContractData.StandaloneMode {
			logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("config.GetGlobalRuntimeConfig().TonMaxMainValidators begin")
			maxValidators, err := config.GetGlobalRuntimeConfig().TonMaxMainValidators(context.Background())
			logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("config.GetGlobalRuntimeConfig().TonMaxMainValidators end")
			if err != nil {
				return nil, err
			}

			info.ValidatorsCountMax = maxValidators
		} else {
			info.ValidatorsCountMax = info.VSetSize
		}

		info.ValidatorsCountNotInDkg = info.ValidatorsCountMax - info.ValidatorsCountInDkg
		info.ValidatorsIdxInDkg = make(map[int]string)
		info.ValidatorsIdxNotInDkg = make(map[int]string)
		info.ValidatorsIdxEvicted = make(map[int]string)

		count := min(info.VSetSize, info.ValidatorsCountMax)

		for i := 0; i < count; i++ {
			pubkey := dkg.VSet[uint16(i)]
			pubkeyBase64 := base64.StdEncoding.EncodeToString(pubkey)

			if dkg.VSetMask.Bit(i) == 1 {
				if dkg.R1.Mask.Bit(i) == 1 {
					info.ValidatorsIdxInDkg[i] = pubkeyBase64
				} else {
					info.ValidatorsIdxNotInDkg[i] = pubkeyBase64
				}
			} else {
				info.ValidatorsIdxEvicted[i] = pubkeyBase64
			}
		}

		info.ValidatorsCountEvicted = len(info.ValidatorsIdxEvicted)

		logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("sumarizeDkgInfo end")

		return &info, nil
	}

	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("sumarizeDkgInfo(status.Original.Dkg) begin")
	dkgInfo, err := sumarizeDkgInfo(status.Original.Dkg)
	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("sumarizeDkgInfo(status.Original.Dkg) end")
	if err != nil {
		return nil, err
	}
	status.DkgInfo = *dkgInfo

	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("sumarizeDkgInfo(status.Original.PrevDkg) begin")
	prevDkgInfo, err := sumarizeDkgInfo(status.Original.PrevDkg)
	logger.Log.Debug().Str("component", "DKG_DEBUG").Msg("sumarizeDkgInfo(status.Original.PrevDkg) end")
	if err != nil {
		return nil, err
	}
	status.PrevDkgInfo = *prevDkgInfo

	return &status, nil
}
