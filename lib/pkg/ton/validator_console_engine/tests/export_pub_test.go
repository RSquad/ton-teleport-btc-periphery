package tests

import (
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/validator_console_engine"
	consoleExecutorMocks "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils/console_executor/mocks"
	"github.com/stretchr/testify/require"
)

func TestExportPub(t *testing.T) {
	exportPubOutput := `connecting to [127.0.0.1:4441]
		local key: C0E4A79C30225D8A9864AE03806A0FBF4BD1F988E889C191D1BD78D8681CB0F1
		remote key: 40C90E768C026594EFABA6A190C89AC4DB0B5F08EFC72CA06FD7E7E949A068F1
		conn ready
		got public key: xrQTSBO18BUkMzZHEXqVAttydp0aQEl6Z7BCpssh2FeLvxWW`
	mc := minimock.NewController(t)

	executorMock := consoleExecutorMocks.NewConsoleExecutorInterfaceMock(mc)
	executorMock.ExecuteMock.Return(exportPubOutput, nil)

	defer t.Cleanup(executorMock.MinimockFinish)

	vce := validator_console_engine.NewValidatorEngineConsole(
		"", "", "", "", executorMock)

	publicKey := vce.ExportPub("DA46DE8CCCED9AB6F29447B334636FBE07F7F4CAE6B6833D26AF1240A1BB34B1")

	expectedPublicKey := "xrQTSBO18BUkMzZHEXqVAttydp0aQEl6Z7BCpssh2FeLvxWW"

	require.Equal(t, publicKey, expectedPublicKey)
}
