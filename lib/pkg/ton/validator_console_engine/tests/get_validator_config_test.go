package tests

import (
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/validator_console_engine"
	consoleExecutorMocks "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils/console_executor/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetValidatorConfig(t *testing.T) {
	getConfigOutput := `connecting to [127.0.0.1:4441]
		local key: C0E4A79C30225D8A9864AE03806A0FBF4BD1F988E889C191D1BD78D8681CB0F1
		remote key: 40C90E768C026594EFABA6A190C89AC4DB0B5F08EFC72CA06FD7E7E949A068F1
		conn ready
		---------
		{
		   "@type" : "engine.validator.config",
		   "validators" : [
			  {
				 "@type" : "engine.validator",
				 "id" : "EZdGkyd0hS7bmxXMxkxztBTrsKr/BWBO3JNnkyWejs0=",
				 "temp_keys" : [
					{
					   "@type" : "engine.validatorTempKey",
					   "key" : "EZdGkyd0hS7bmxXMxkxztBTrsKr/BWBO3JNnkyWejs0=",
					   "expire_at" : 1730499055
					}
				 ],
				 "adnl_addrs" : [
					{
					   "@type" : "engine.validatorAdnlAddress",
					   "id" : "3sBXy7GxE/4hfL3C2AuSTls+ZBSywW2uELbYLly2Q34=",
					   "expire_at" : 1730499055
					}
				 ],
				 "election_date" : 1730498455,
				 "expire_at" : 1730499055
			  }
		   ]
		}
		---------`
	mc := minimock.NewController(t)

	executorMock := consoleExecutorMocks.NewConsoleExecutorInterfaceMock(mc)
	// executorMock.ExecuteMock.When("/validator-engine-console -p  -k  -a  -v 0 -c \"getconfig\"").Then(getConfigOutput, nil)
	executorMock.ExecuteMock.Expect("/validator-engine-console -p  -k  -a  -v 0 -c \"getconfig\"").Return(getConfigOutput, nil)

	defer t.Cleanup(executorMock.MinimockFinish)

	vce := validator_console_engine.NewValidatorEngineConsole(
		"", "", "", "", executorMock)

	validatorConfig := vce.GetValidatorConfig()

	require.Equal(t, validatorConfig.Validators, &validator_console_engine.ValidatorEngineConfig{})
	expectedConfig := &validator_console_engine.ValidatorEngineConfig{
		Validators: []validator_console_engine.Validator{
			{
				ID:           "EZdGkyd0hS7bmxXMxkxztBTrsKr/BWBO3JNnkyWejs0=",
				ElectionDate: 1730498455,
				ExpireAt:     1730499055,
			},
		},
	}

	require.Equal(t, validatorConfig.Validators, expectedConfig.Validators)
}
