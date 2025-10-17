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

type BtcTxStatus int

const (
	BTC_TX_NOT_PUBLISHED BtcTxStatus = iota
	BTC_TX_IN_MEMPOOL    BtcTxStatus = 1
	BTC_TX_IN_BLOCK      BtcTxStatus = 2
)

type SysLastPegoutTxInfo struct {
	BtcTxStatus        BtcTxStatus
	BitcoinMempoolTime int
	CPFP               int
}

type SigningStatus int

const (
	SIGNING_STATUS_NOT_SIGNED SigningStatus = iota
	SIGNING_STATUS_SIGNED
	NO_PEGOUT = -1
)

type SysPegoutSigningInfo struct {
	Id                           int
	Restarts                     int
	RestartsSeverity             alerts.Severity
	Until                        time.Time
	QueueLength                  int
	IsSigned                     SigningStatus
	IsAutopegout                 bool
	Signers                      int
	SignersMax                   int
	SignersCommitmentActive      int
	SignersCommitmentActiveIdx   []int
	SignersCommitmentInactive    int
	SignersCommitmentInactiveIdx []int
	SignersEvicted               int
	SignersEvictedIdx            []int
	IsInternalKeyCorrect         bool
	IsInternalKeyCorrectStr      string
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
