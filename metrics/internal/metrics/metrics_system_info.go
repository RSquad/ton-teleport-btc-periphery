package metrics

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type MetricsSystemInfo struct {
}

func (systemInfo *MetricsSystemInfo) SystemInfoJson(
	dataSourceDB *data_sources.DataSourceDB,
	alertManager *alerts.AlertManager,
	contractAddrs map[string]*address.Address,
	globalRuntimeConfig *config.GlobalRuntimeConfig,
	bitcoinClient *bitcoin.Client,
) (string, error) {
	sysDkgInfo, err := systemInfo.SysDkgInfo(dataSourceDB, alertManager, contractAddrs, globalRuntimeConfig)
	if err != nil {
		return "", err
	}

	sysLastPegoutTxInfo, err := systemInfo.SysLastPegoutTxInfo(dataSourceDB, bitcoinClient)
	if err != nil {
		return "", err
	}

	sysPegoutSigningInfo, err := systemInfo.PegoutSigningInfo(dataSourceDB)
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
	globalRuntimeConfig *config.GlobalRuntimeConfig,
) (*SysDkgInfo, error) {
	dkg, err := dataSourceDB.Dkg()
	if err != nil {
		return nil, err
	}

	// restarts, restartsSeverity
	dkgRestartsAlertState, err := alertManager.GetAlertState("alert_dkg_restarts")
	if err != nil {
		return nil, err
	}

	// culprits, culpritsSeverity
	dkgCulpritsAlertState, err := alertManager.GetAlertState("alert_dkg_culprit_found")
	if err != nil {
		return nil, err
	}

	// DkgStatus
	dkgStatus, err := systemInfo.DkgStatus(dataSourceDB, globalRuntimeConfig)
	if err != nil {
		return nil, err
	}

	return &SysDkgInfo{
		State:            dkg.State,
		StateName:        dkg.State.String(),
		Restarts:         int(dkgRestartsAlertState.Values["restarts"].(int64)),
		RestartsSeverity: dkgRestartsAlertState.Severity,
		CulpritsIdx:      dkgCulpritsAlertState.Values["culprit_id"].([]int),
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
	}, nil
}

func (systemInfo *MetricsSystemInfo) SysLastPegoutTxInfo(
	dataSourceDB *data_sources.DataSourceDB,
	bitcoinClient *bitcoin.Client,
) (*SysLastPegoutTxInfo, error) {
	lastSignedPegout, err := dataSourceDB.LastSignedPegout()
	if err != nil {
		return nil, err
	}

	// BtcTxStatus
	btcTxStatus := BTC_TX_NOT_PUBLISHED
	btcTxTimestamp := int64(0)
	{
		{
			btcMempoolEntry, err := bitcoinClient.RPCClient.GetMempoolEntry(mutils.BytesToBTCHash(lastSignedPegout.BitcoinTxId).String())
			if err == nil {
				if btcMempoolEntry != nil {
					btcTxTimestamp = btcMempoolEntry.Time
					btcTxStatus = BTC_TX_IN_MEMPOOL
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
	}

	// BitcoinMempoolTime
	bitcoinMempoolTime := int64(0)
	if btcTxStatus == BTC_TX_IN_MEMPOOL {
		bitcoinMempoolTime = int64(time.Duration(time.Now().Unix()-btcTxTimestamp) * time.Second)
	}

	return &SysLastPegoutTxInfo{
		BtcTxStatus:        btcTxStatus,
		BitcoinMempoolTime: int(bitcoinMempoolTime),
		CPFP:               -1, // TODO: implement CPFP
	}, nil
}

func (systemInfo *MetricsSystemInfo) PegoutSigningInfo(
	dataSourceDB *data_sources.DataSourceDB,
) (*SysPegoutSigningInfo, error) {

	// TODO: implement

	return &SysPegoutSigningInfo{
		Id:                           123,
		Restarts:                     0,
		CulpritsIdx:                  make([]int, 0),
		Until:                        time.Now(),
		Signers:                      89,
		QueueLength:                  3,
		IsSigned:                     true,
		IsSignedStr:                  "Yes",
		SignersMax:                   100,
		SignersCommitmentActive:      89,
		SignersCommitmentActiveIdx:   make([]int, 89),
		SignersCommitmentInactive:    19,
		SignersCommitmentInactiveIdx: make([]int, 19),
		SignersEvicted:               0,
		SignersEvictedIdx:            make([]int, 0),
		IsInternalKeyCorrect:         true,
		IsInternalKeyCorrectStr:      "Internal Key Is Correct",
	}, nil
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

	// TODO: implement

	return &SysTeleportInfo{
		UTXO:                       0,
		IsSameInputInternalKey:     true,
		IsSameInputInternalKeyStr:  "The same input internal key",
		TimeSinceLastAutopegout:    300,
		TimeSinceLastAutopegoutStr: "5 min",
		ServiceFee:                 -140,
		LastConfirmed:              0,
		LastBtc_LastTon:            3,
		LastTon_PegoutBlock:        45,
	}, nil
}

func (systemInfo *MetricsSystemInfo) DkgStatus(
	dataSourceDB *data_sources.DataSourceDB,
	globalRuntimeConfig *config.GlobalRuntimeConfig,
) (*DkgStatus, error) {
	dkg, err := dataSourceDB.Dkg()
	if err != nil {
		return nil, err
	}

	lastRestartDkg, err := dataSourceDB.DkgBeforeRestart(dkg.Until)
	if err != nil {
		return nil, err
	}

	prevDkg, err := dataSourceDB.PrevDkg()
	if err != nil {
		return nil, err
	}

	coordinatorContractData, err := dataSourceDB.CoordinatorContractStorage()
	if err != nil {
		return nil, err
	}

	var status DkgStatus
	status.StandaloneMode = coordinatorContractData.StandaloneMode
	status.Original.Dkg = dkg
	status.Original.LastRestartDkg = lastRestartDkg
	status.Original.PrevDkg = prevDkg

	sumarizeDkgInfo := func(dkg *coordinator.DKG) (*DkgInfo, error) {
		var info DkgInfo

		info.State = dkg.State.String()
		info.VSetSize = len(dkg.VSet)

		if dkg.R1 != nil {
			info.ValidatorsCountInDkg = int(dkg.R1.Count)
		} else {
			info.ValidatorsCountInDkg = 0
		}

		if !coordinatorContractData.StandaloneMode {
			maxValidators, err := globalRuntimeConfig.TonMaxMainValidators(context.Background()) // TODO: replace with user defined ctx
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

		return &info, nil
	}

	dkgInfo, err := sumarizeDkgInfo(status.Original.Dkg)
	if err != nil {
		return nil, err
	}
	status.DkgInfo = *dkgInfo

	prevDkgInfo, err := sumarizeDkgInfo(status.Original.PrevDkg)
	if err != nil {
		return nil, err
	}
	status.PrevDkgInfo = *prevDkgInfo

	return &status, nil
}
