package metrics

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/xssnick/tonutils-go/address"
)

type DkgOriginalData struct {
	Dkg            *coordinator.DKG
	LastRestartDkg *coordinator.DKG
	PrevDkg        *coordinator.DKG
}

type DkgInfo struct {
	State              string
	VSetSize           int
	ValidatorsCountMax int

	ValidatorsCountInDkg    int
	ValidatorsCountNotInDkg int
	ValidatorsCountEvicted  int

	ValidatorsIdxInDkg    map[int]string
	ValidatorsIdxNotInDkg map[int]string
	ValidatorsIdxEvicted  map[int]string
}

type DkgStatus struct {
	StandaloneMode bool
	DkgInfo        DkgInfo
	PrevDkgInfo    DkgInfo
	Original       DkgOriginalData
}

type SysDkgInfo struct {
	State                 coordinator.DKGState
	StateName             string
	Restarts              int
	RestartsSeverity      alerts.Severity
	CulpritsIdx           []int
	CulpritsSeverity      alerts.Severity
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

type SysLastPegoutTxInfo struct {
	IsInternalKeyCorrect    bool
	IsInternalKeyCorrectStr string
	IsSigned                bool
	IsSignedStr             string
	BitcoinMempool          int
	CPFP                    int
}

type SysPegoutSigningInfo struct {
	Id                           int
	Restarts                     int
	CulpritsIdx                  []int
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

type SysBalancesInfo struct {
	Coordinator         int64
	CoordinatorStr      string
	CoordinatorAddr     *address.Address
	CoordinatorSeverity alerts.Severity
	Teleport            int64
	TeleportStr         string
	TeleportAddr        *address.Address
	TeleportSeverity    alerts.Severity
	Bitclient           int64
	BitclientStr        string
	BitclientAddr       *address.Address
	BitclientSeverity   alerts.Severity
	Minter              int64
	MinterStr           string
	MinterAddr          *address.Address
	MinterSeverity      alerts.Severity
	Relayer             int64
	RelayerStr          string
	RelayerAddr         *address.Address
	RelayerSeverity     alerts.Severity
}

type SysTeleportInfo struct {
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
	DkgInfo           *SysDkgInfo
	LastPegoutTxInfo  *SysLastPegoutTxInfo
	PegoutSigningInfo *SysPegoutSigningInfo
	BalancesInfo      *SysBalancesInfo
	TeleportInfo      *SysTeleportInfo
}
