package coordinatorcontract

import (
	"fmt"
	"strings"
)

const (
	ErrorDKGInProgress      = 126
	ErrorDKGPart1Incomplete = 127
	ErrorDKGPart2Incomplete = 128
)

func (c *CoordinatorContract) containsErrorCode(err error, errorCode int) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("exitcode=%d", errorCode))
}

func (c *CoordinatorContract) IsDKGInProgressError(err error) bool {
	return c.containsErrorCode(err, ErrorDKGInProgress)
}

func (c *CoordinatorContract) IsDKGPart1IncompleteError(err error) bool {
	return c.containsErrorCode(err, ErrorDKGPart1Incomplete)
}

func (c *CoordinatorContract) IsDKGPart2IncompleteError(err error) bool {
	return c.containsErrorCode(err, ErrorDKGPart2Incomplete)
}
