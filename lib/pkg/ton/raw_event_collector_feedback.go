package ton

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

func (ec *RawEventCollector) logStartWork() {
	logger.Log.Info().
		Str("component", "RawEventCollector").
		Str("addr", utils.AddrToRawString(ec.addr)).
		Msg("Start collecting events")
}

func (ec *RawEventCollector) logFinishWork(err error) {
	event := logger.Log.Info().
		Str("component", "RawEventCollector").
		Str("addr", utils.AddrToRawString(ec.addr))

	if err != nil {
		event.Err(err).Msg("Finished collecting events with error")
	} else {
		event.Msg("Finished collecting events")
	}
}
