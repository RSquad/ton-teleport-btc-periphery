package alerts

import "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"

func logNoTeleportStorage(component string) {
	logger.Log.Debug().
		Str("component", component).
		Msg("Teleport contract storage not found")
}
