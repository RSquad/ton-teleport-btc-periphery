package metrics

import (
	"encoding/json"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type MetricsSystemInfo struct {
}

type DkgInfo struct {
	State                 coordinator.DKGState
	StateName             string
	Restarts              int
	Culprits              string
	Until                 time.Time
	ValidatorsMax         int
	ValidatorsActive      int
	ValidatorsActiveIdx   []int
	ValidatorsInactive    int
	ValidatorsInactiveIdx []int
	ValidatorsEvicted     int
	ValidatorsEvictedIdx  []int
	Timeout               int
}

type LastPegoutTxInfo struct {
	IsInternalKeyCorrect    bool
	IsInternalKeyCorrectStr string
	IsSigned                bool
	IsSignedStr             string
	BitcoinMempool          int
	CPFP                    int
}

type PegoutSigningInfo struct {
	Id                           int
	Restarts                     int
	Culprits                     string
	Until                        time.Time
	Signers                      int
	QueueLength                  int
	IsSigned                     bool
	IsSignedStr                  string
	SignersMax                   int
	SignersCommitmentActive      int
	SignersCommitmentActiveIdx   []int
	SignersCommitmentInactive    int
	SignersCommitmentInactiveIdx []int
	SignersEvicted               int
	SignersEvictedIdx            []int
}

type BalancesInfo struct {
	Coordinator         int64
	CoordinatorStr      string
	CoordinatorAddr     *address.Address
	CoordinatorSeverity int
	Teleport            int64
	TeleportStr         string
	TeleportAddr        *address.Address
	TeleportSeverity    int
	Bitclient           int64
	BitclientStr        string
	BitclientAddr       *address.Address
	BitclientSeverity   int
	Minter              int64
	MinterStr           string
	MinterAddr          *address.Address
	MinterSeverity      int
	Relayer             int64
	RelayerStr          string
	RelayerAddr         *address.Address
	RelayerSeverity     int
}

type TeleportInfo struct {
	UTXO                       int
	IsSameInputInternalKey     bool
	IsSameInputInternalKeyStr  string
	TimeSinceLastAutopegout    int // seconds
	TimeSinceLastAutopegoutStr string
	ServiceFee                 int
	LastConfirmed              int
	LastBtc_LastTon            int
	LastTon_PegoutBlock        int
}

type SystemInfo struct {
	DkgInfo           *DkgInfo
	LastPegoutTxInfo  *LastPegoutTxInfo
	PegoutSigningInfo *PegoutSigningInfo
	BalancesInfo      *BalancesInfo
	TeleportInfo      *TeleportInfo
}

func (systemInfo *MetricsSystemInfo) SystemInfoJson(
	dataSourceDB *data_sources.DataSourceDB,
	alertManager *alerts.AlertManager,
	contractAddrs map[string]*address.Address,
) (string, error) {
	balancesInfo, err := systemInfo.BalancesInfo(dataSourceDB, alertManager, contractAddrs)
	if err != nil {
		return "", err
	}

	info := SystemInfo{
		DkgInfo:           nil,
		LastPegoutTxInfo:  nil,
		PegoutSigningInfo: nil,
		BalancesInfo:      balancesInfo,
		TeleportInfo:      nil,
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (systemInfo *MetricsSystemInfo) BalancesInfo(
	dataSourceDB *data_sources.DataSourceDB,
	alertManager *alerts.AlertManager,
	contractAddrs map[string]*address.Address,
) (*BalancesInfo, error) {
	coordinator, err := dataSourceDB.ActualContractBalance("coordinator")
	if err != nil {
		return nil, err
	}
	coordinatorAlertState, err := alertManager.GetAlertState("contract_balance_coordinator")
	if err != nil {
		return nil, err
	}

	teleport, err := dataSourceDB.ActualContractBalance("teleport")
	if err != nil {
		return nil, err
	}
	teleportAlertState, err := alertManager.GetAlertState("contract_balance_teleport")
	if err != nil {
		return nil, err
	}

	bitclient, err := dataSourceDB.ActualContractBalance("bitclient")
	if err != nil {
		return nil, err
	}
	bitclientAlertState, err := alertManager.GetAlertState("contract_balance_bitclient")
	if err != nil {
		return nil, err
	}

	minter, err := dataSourceDB.ActualContractBalance("minter")
	if err != nil {
		return nil, err
	}
	minterAlertState, err := alertManager.GetAlertState("contract_balance_minter")
	if err != nil {
		return nil, err
	}

	relayer, err := dataSourceDB.ActualContractBalance("relayer")
	if err != nil {
		return nil, err
	}
	relayerAlertState, err := alertManager.GetAlertState("contract_balance_relayer")
	if err != nil {
		return nil, err
	}

	return &BalancesInfo{
		Coordinator:         coordinator,
		CoordinatorStr:      mutils.NanoIntToString(coordinator),
		CoordinatorSeverity: int(coordinatorAlertState.Severity),
		CoordinatorAddr:     contractAddrs["coordinator"],
		Teleport:            teleport,
		TeleportStr:         mutils.NanoIntToString(teleport),
		TeleportAddr:        contractAddrs["teleport"],
		TeleportSeverity:    int(teleportAlertState.Severity),
		Bitclient:           bitclient,
		BitclientStr:        mutils.NanoIntToString(bitclient),
		BitclientAddr:       contractAddrs["bitclient"],
		BitclientSeverity:   int(bitclientAlertState.Severity),
		Minter:              minter,
		MinterStr:           mutils.NanoIntToString(minter),
		MinterAddr:          contractAddrs["minter"],
		MinterSeverity:      int(minterAlertState.Severity),
		Relayer:             relayer,
		RelayerStr:          mutils.NanoIntToString(relayer),
		RelayerAddr:         contractAddrs["relayer"],
		RelayerSeverity:     int(relayerAlertState.Severity),
	}, nil
}
