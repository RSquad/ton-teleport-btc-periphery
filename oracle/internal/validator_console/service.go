package validatorconsole

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

type ValidatorTempKey struct {
	Key      string `json:"key"`
	ExpireAt int64  `json:"expire_at"`
}

type ValidatorAdnlAddress struct {
	Id       string `json:"id"`
	ExpireAt int64  `json:"expire_at"`
}

type Validator struct {
	Id           string                 `json:"id"`
	TempKeys     []ValidatorTempKey     `json:"temp_keys"`
	AdnlAddrs    []ValidatorAdnlAddress `json:"adnl_addrs"`
	ElectionDate int64                  `json:"election_date"`
	ExpireAt     int64                  `json:"expire_at"`
}

type ValidatorEngineConfig struct {
	OutPort    int64
	Validators []Validator
}

type ValidatorConsole struct {
	validatorEngineConsolePath string
	serverPublicKeyPath        string
	clientPrivateKeyPath       string
	serverAddress              string
}

type ValidatorConsoleKey struct {
	ValidatorKey []byte
	ValidatorId  string
}

func NewValidatorConsole(
	validatorEngineConsolePath string,
	serverPublicKeyPath string,
	clientPrivateKeyPath string,
	serverAddress string,
) *ValidatorConsole {
	return &ValidatorConsole{
		validatorEngineConsolePath: validatorEngineConsolePath,
		serverPublicKeyPath:        serverPublicKeyPath,
		clientPrivateKeyPath:       clientPrivateKeyPath,
		serverAddress:              serverAddress,
	}
}

func (v *ValidatorConsole) runCommand(c string) ([]byte, error) {
	args := v.buildCommand(c)
	cmd := exec.Command(v.validatorEngineConsolePath+"/validator-engine-console", args...)

	return cmd.CombinedOutput()
}

func (v *ValidatorConsole) buildCommand(c string) []string {
	return []string{
		"-c", c,
		"-p", v.serverPublicKeyPath,
		"-k", v.clientPrivateKeyPath,
		"-a", v.serverAddress,
		"-v", "\"0\"",
	}
}

func (v *ValidatorConsole) GetValidatorKeys() ([]ValidatorConsoleKey, error) {
	command, err := v.runCommand("getconfig")
	if err != nil {
		return nil, err
	}

	fixedStr, err := extractJSON(string(command))
	if err != nil {
		return nil, err
	}

	var config ValidatorEngineConfig
	if err := json.Unmarshal([]byte(fixedStr), &config); err != nil {
		return nil, err
	}

	validatorConsoleKeys := make([]ValidatorConsoleKey, 0, len(config.Validators))
	for _, validator := range config.Validators {
		pubKey, err := v.exportPub(validator.Id)
		if err != nil {
			return nil, err
		}

		base64Id, err := base64.StdEncoding.DecodeString(validator.Id)
		if err != nil {
			return nil, err
		}

		validatorConsoleKey := ValidatorConsoleKey{
			ValidatorKey: pubKey,
			ValidatorId:  hex.EncodeToString(base64Id),
		}
		validatorConsoleKeys = append(validatorConsoleKeys, validatorConsoleKey)
	}
	return validatorConsoleKeys, nil
}

func extractJSON(input string) (string, error) {
	re := regexp.MustCompile(`-+[\r\n]+([\s\S]*?)[\r\n]+-+`)
	match := re.FindString(input)
	if match == "" {
		return "", fmt.Errorf("no JSON found in input")
	}
	return strings.Trim(strings.TrimSpace(match), "-"), nil
}

func extractPublicKey(output string) string {
	regex := regexp.MustCompile(`got public key:\s*([A-Za-z0-9+/=]+)`)
	match := regex.FindString(output)
	match = strings.Trim(match, "got public key: ")
	return match
}

func (v *ValidatorConsole) exportPub(validatorIdBase64 string) ([]byte, error) {
	p, err := base64.StdEncoding.DecodeString(validatorIdBase64)
	if err != nil {
		return nil, err
	}

	validatorKey := hex.EncodeToString(p)
	result, err := v.runCommand(fmt.Sprintf("exportpub %s", validatorKey))
	if err != nil {
		return nil, err
	}
	base64Result := extractPublicKey(string(result))
	d, err := base64.StdEncoding.DecodeString(base64Result)

	if err != nil {
		return nil, err
	}

	return d, nil
}

func extractSignature(output string) string {
	regex := regexp.MustCompile(`got signature\s+([A-Za-z0-9+/=]+)`)
	match := regex.FindString(output)
	match = strings.Trim(match, "got signature: ")
	return match
}

func (v *ValidatorConsole) Sign(validatorId string, hash string) (string, error) {
	result, err := v.runCommand(fmt.Sprintf("sign %s %s", validatorId, hash))
	if err != nil {
		return "", err
	}
	return extractSignature(string(result)), nil
}

type ValidatorSigner struct {
	validatorEngineCondole ValidatorConsole
	publicKey              string
}

func (v *ValidatorConsole) NewValidatorSigner(publicKey string) *ValidatorSigner {
	return &ValidatorSigner{validatorEngineCondole: *v, publicKey: publicKey}
}

func (s *ValidatorSigner) SignCell(cell *cell.Cell) []byte {
	cellHex := hex.EncodeToString(cell.Hash())
	result, err := s.validatorEngineCondole.Sign(s.publicKey, cellHex)
	if err != nil {
		return nil
	}

	return []byte(result)
}
