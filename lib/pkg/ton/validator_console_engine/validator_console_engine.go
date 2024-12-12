package validator_console_engine

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils/console_executor"
)

const DefaultVerbosityLevel = 0

type Validator struct {
	ID           string `json:"id"`
	ElectionDate int    `json:"election_date"`
	ExpireAt     int    `json:"expire_at"`
}

type ValidatorEngineConfig struct {
	Validators []Validator `json:"validators"`
}

type ValidatorKeysResponse struct {
	ValidatorKeys []string
	ValidatorIds  []string
}

type BaseCommandParams struct {
	P string // Server public key path
	K string // Client private key path
	A string // Server address
	V int    // Verbosity level
}

type ValidatorEngineCommand struct {
	BaseCommandParams
	C string
}

type ValidatorEngineConsole struct {
	ValidatorEngineConsolePath string
	ServerPublicKeyPath        string
	ClientPrivateKeyPath       string
	ServerAddress              string
	baseCommandParams          BaseCommandParams
	consoleExecutor            console_executor.ConsoleExecutorInterface
}

func NewValidatorEngineConsole(
	validatorEngineConsolePath string,
	serverPublicKeyPath string,
	clientPrivateKeyPath string,
	serverAddress string,
	executor console_executor.ConsoleExecutorInterface,
) *ValidatorEngineConsole {
	baseCommandParams := BaseCommandParams{
		P: serverPublicKeyPath,
		K: clientPrivateKeyPath,
		A: serverAddress,
		V: DefaultVerbosityLevel,
	}
	return &ValidatorEngineConsole{
		ValidatorEngineConsolePath: validatorEngineConsolePath,
		ServerPublicKeyPath:        serverPublicKeyPath,
		ClientPrivateKeyPath:       clientPrivateKeyPath,
		ServerAddress:              serverAddress,
		baseCommandParams:          baseCommandParams,
		consoleExecutor:            executor,
	}
}

func (v *ValidatorEngineConsole) GetValidatorConfig() *ValidatorEngineConfig {
	params := ValidatorEngineCommand{
		BaseCommandParams: v.baseCommandParams,
		C:                 "getconfig",
	}

	result, err := v.runCommand(params)
	if err != nil {
		log.Printf("Error running command getconfig: %v", err)
	}

	config, _ := v.parseGetValidatorConfigOutput(result)
	return config
}

func (v *ValidatorEngineConsole) ExportPub(validatorIdBase64 string) string {
	command := fmt.Sprintf("exportpub %x", validatorIdBase64)

	params := ValidatorEngineCommand{
		BaseCommandParams: v.baseCommandParams,
		C:                 command,
	}
	result, err := v.runCommand(params)
	if err != nil {
		log.Printf("Error running command exportpub: %v", err)
	}

	publicKey, _ := v.parseExportPubOutput(result)

	return publicKey
}

func (v *ValidatorEngineConsole) GetValidatorKeys() *ValidatorKeysResponse {
	validatorsConfig := v.GetValidatorConfig()
	var validatorKeys []string
	var validatorIds []string
	for _, val := range validatorsConfig.Validators {
		valId := val.ID
		publicKey := v.ExportPub(valId)
		if len(publicKey) > 8 {
			validatorKeys = append(validatorKeys, publicKey[8:])
		} else {
			validatorKeys = append(validatorKeys, publicKey)
		}
		valHexId := fmt.Sprintf("%x", valId)
		validatorIds = append(validatorIds, valHexId)
	}
	return &ValidatorKeysResponse{
		ValidatorKeys: validatorKeys,
		ValidatorIds:  validatorIds,
	}
}

func (v *ValidatorEngineConsole) GetValidatorPublicKey(timestamp int64) (string, error) {
	config := v.GetValidatorConfig()

	validators := config.Validators
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].ElectionDate > validators[j].ElectionDate
	})

	for _, validator := range validators {
		validatorId := validator.ID
		validatorKey := fmt.Sprintf("%x", validatorId)
		validatorKey = strings.ToUpper(validatorKey)

		if int64(validator.ElectionDate) < timestamp && timestamp < int64(validator.ExpireAt) {
			return validatorKey, nil
		}
	}
	return "", fmt.Errorf("GetValidatorKey error: validator key not found.")
}

func (v *ValidatorEngineConsole) runCommand(params ValidatorEngineCommand) (string, error) {
	executable := fmt.Sprintf("%s/validator-engine-console", v.ValidatorEngineConsolePath)
	command := v.buildCommand(executable, params)
	if v.consoleExecutor == nil {
		return "", fmt.Errorf("executor is nil")
	}
	result, err := v.consoleExecutor.Execute(command)
	if err != nil {
		log.Fatalf("Error running command: %v", err)
	}

	return result, nil
}

func (v *ValidatorEngineConsole) buildCommand(executable string, parameters ValidatorEngineCommand) string {
	return fmt.Sprintf("%s -p %s -k %s -a %s -v %d -c \"%s\"",
		executable,
		parameters.P,
		parameters.K,
		parameters.A,
		parameters.V,
		parameters.C,
	)
}

func (v *ValidatorEngineConsole) parseGetValidatorConfigOutput(output string) (*ValidatorEngineConfig, error) {
	re := regexp.MustCompile(`(?s)-+\s*\n(.*?)\n-+`)
	match := re.FindStringSubmatch(output)

	if len(match) < 2 {
		return nil, fmt.Errorf("Error parsing JSON from output")
	}
	jsonStr := strings.TrimSpace(match[1])

	var config ValidatorEngineConfig
	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling JSON: %v", err)
	}
	return &config, nil
}

func (v *ValidatorEngineConsole) parseExportPubOutput(output string) (string, error) {
	re := regexp.MustCompile(`got public key:\s*([A-Za-z0-9+/=]+)`)
	match := re.FindStringSubmatch(output)

	if len(match) < 2 {
		return "", fmt.Errorf("Error parsing public key from output")
	}

	return match[1], nil
}
