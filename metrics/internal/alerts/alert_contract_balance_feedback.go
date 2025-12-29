package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

func logBalanceFetchError(alert *AlertContractBalance, err error) {
	logger.Log.Error().
		Str("component", "AlertContractBalance").
		Str("alert_name", alert.Name).
		Str("balance_name", alert.BalanceName).
		Str("contract_address", utils.AddrToRawString(alert.Addr)).
		Err(err).
		Msg("Failed to fetch contract balance")
}

func logLowBalanceAlert(alert *AlertContractBalance, severity Severity, balance int64) {
	balanceStr := mutils.NanoIntToString(balance)

	logger.Log.Warn().
		Str("component", "AlertContractBalance").
		Str("alert_name", alert.Name).
		Str("balance_name", alert.BalanceName).
		Str("contract_address", utils.AddrToRawString(alert.Addr)).
		Str("severity", severity.String()).
		Str("balance_ton", balanceStr).
		Int64("balance_nano", balance).
		Msg("Contract low balance alert triggered")
}

func logBalanceCheckPassed(alert *AlertContractBalance, balance int64) {
	balanceStr := mutils.NanoIntToString(balance)

	logger.Log.Debug().
		Str("component", "AlertContractBalance").
		Str("alert_name", alert.Name).
		Str("balance_name", alert.BalanceName).
		Str("contract_address", utils.AddrToRawString(alert.Addr)).
		Str("balance_ton", balanceStr).
		Int64("balance_nano", balance).
		Msg("Contract balance check passed")
}
