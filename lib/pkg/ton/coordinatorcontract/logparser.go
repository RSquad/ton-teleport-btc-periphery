package coordinatorcontract

import (
	"fmt"
	"time"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	logIdDKGComplete = 0x453443a6
)

type LogInterface interface {
	GetLogID() uint32
}

type DKGCompletedLog struct {
	CompletedAt time.Time
	Key         []byte
}

func (m *DKGCompletedLog) GetLogID() uint32 {
	return logIdDKGComplete
}

type LogParser struct{}

func NewLogParser() (
	*LogParser,
	error,
) {
	return &LogParser{}, nil
}

func (c *LogParser) Parse(logCell *cell.Cell) (LogInterface, error) {
	logSlice := logCell.BeginParse()
	logId := logSlice.MustLoadUInt(32)

	switch logId {
	case logIdDKGComplete:
		mintLog, err := parseDKGCompleteLog(logSlice)
		return mintLog, err
	default:
		return nil, fmt.Errorf("[LogParser] unknown log type with log id %x", logId)
	}
}

func parseDKGCompleteLog(logSlice *cell.Slice) (*DKGCompletedLog, error) {
	completedAt := logSlice.MustLoadBigUInt(64)
	key := logSlice.MustLoadSlice(256)
	return &DKGCompletedLog{
		CompletedAt: time.Unix(completedAt.Int64(), 0),
		Key:         key,
	}, nil
}
