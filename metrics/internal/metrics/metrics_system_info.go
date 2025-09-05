package metrics

import (
	"encoding/json"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
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
	CoordinatorSeverity int
	Teleport            int64
	TeleportStr         string
	TeleportSeverity    int
	Bitclient           int64
	BitclientStr        string
	BitclientSeverity   int
	Minter              int64
	MinterStr           string
	MinterSeverity      int
	Relayer             int64
	RelayerStr          string
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

func (systemInfo *MetricsSystemInfo) SystemInfoJson(dataSourceDB *data_sources.DataSourceDB) (string, error) {
	balancesInfo, err := systemInfo.BalancesInfo(dataSourceDB)
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

func (systemInfo *MetricsSystemInfo) BalancesInfo(dataSourceDB *data_sources.DataSourceDB) (*BalancesInfo, error) {
	coordinator, err := dataSourceDB.ActualContractBalance("coordinator")
	if err != nil {
		return nil, err
	}

	teleport, err := dataSourceDB.ActualContractBalance("teleport")
	if err != nil {
		return nil, err
	}

	bitclient, err := dataSourceDB.ActualContractBalance("bitclient")
	if err != nil {
		return nil, err
	}

	minter, err := dataSourceDB.ActualContractBalance("minter")
	if err != nil {
		return nil, err
	}

	relayer, err := dataSourceDB.ActualContractBalance("relayer")
	if err != nil {
		return nil, err
	}

	return &BalancesInfo{
		Coordinator:         coordinator,
		CoordinatorStr:      "",
		CoordinatorSeverity: 0,
		Teleport:            teleport,
		TeleportStr:         "",
		TeleportSeverity:    0,
		Bitclient:           bitclient,
		BitclientStr:        "",
		BitclientSeverity:   0,
		Minter:              minter,
		MinterStr:           "",
		MinterSeverity:      0,
		Relayer:             relayer,
		RelayerStr:          "",
		RelayerSeverity:     0,
	}, nil
}
