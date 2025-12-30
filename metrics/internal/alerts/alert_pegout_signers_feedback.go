package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func logPegoutSignersAlert(component string, severity Severity, addr *address.Address,
	btcTxId string, signersAllowedCount, maxSigners int, percentage uint,
) {
	logger.Log.Warn().
		Str("component", component).
		Str("severity", severity.String()).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Str("bitcoin_tx_id", btcTxId).
		Int("signers_allowed", signersAllowedCount).
		Int("max_signers", maxSigners).
		Uint("percentage", percentage).
		Msg("Pegout signers alert triggered")
}

func logPegoutSignersCheckPassed(component string, addr *address.Address, btcTxId string,
	signersAllowedCount, maxSigners int, percentage uint,
) {
	logger.Log.Debug().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Str("bitcoin_tx_id", btcTxId).
		Int("signers_allowed", signersAllowedCount).
		Int("max_signers", maxSigners).
		Uint("percentage", percentage).
		Msg("Pegout signers check passed")
}
