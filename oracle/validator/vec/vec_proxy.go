package vec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type Proxy struct {
	vecPath              string
	serverPublicKeyPath  string
	clientPrivateKeyPath string
	validatorServerAddr  string
}

type Validator struct {
	ID           string `json:"id"`
	ElectionDate int64  `json:"election_date"`
	ExpireAt     int64  `json:"expire_at"`
}

type ValidatorEngineConfig struct {
	Validators []Validator `json:"validators"`
}

type ValidatorKeysResponse struct {
	ValidatorKeys []string
	ValidatorIds  []string
}

const DEFAULT_VERBOSITY_LEVEL = "0"

func NewProxy(
	vecPath string,
	serverPublicKeyPath string,
	clientPrivateKeyPath string,
	validatorServerAddr string,
) *Proxy {
	return &Proxy{
		vecPath:              vecPath,
		serverPublicKeyPath:  serverPublicKeyPath,
		clientPrivateKeyPath: clientPrivateKeyPath,
		validatorServerAddr:  validatorServerAddr,
	}
}

func (p *Proxy) Sign(validatorPublicKey, hash string) (string, error) {
	params := p.buildCmd("sign", validatorPublicKey, hash)
	output, err := p.runCmd(params)
	if err != nil {
		return "", err
	}
	return p.extractSignature(output)
}

func (p *Proxy) GetKeys() (*ValidatorKeysResponse, error) {
	cfg, err := p.getCfg()
	if err != nil {
		return nil, err
	}

	var validatorKeys []string
	var validatorIds []string

	for _, val := range cfg.Validators {
		valID, err := p.exportPub(val.ID)
		if err != nil {
			return nil, err
		}
		validatorKeys = append(validatorKeys, valID[8:])
		validatorIds = append(validatorIds, utils.MustBase64ToHexStr(val.ID, 32))
	}

	return &ValidatorKeysResponse{ValidatorKeys: validatorKeys, ValidatorIds: validatorIds}, nil
}

func (p *Proxy) getCfg() (*ValidatorEngineConfig, error) {
	params := p.buildCmd("getconfig")
	output, err := p.runCmd(params)
	if err != nil {
		return nil, err
	}
	return p.extractCfg(output)
}

func (p *Proxy) buildCmd(cmd string, args ...string) []string {
	params := []string{
		fmt.Sprintf("-%s=%s", "p", p.serverPublicKeyPath),
		fmt.Sprintf("-%s=%s", "k", p.clientPrivateKeyPath),
		fmt.Sprintf("-%s=%s", "a", p.validatorServerAddr),
		fmt.Sprintf("-%s=%s", "v", DEFAULT_VERBOSITY_LEVEL),
	}
	params = append(params, fmt.Sprintf("-%s='%s'", "c", strings.Join(append([]string{cmd}, args...), " ")))
	return append([]string{fmt.Sprintf("%s/validator-engine-console", p.vecPath)}, params...)
}

func (v *Proxy) runCmd(params []string) (string, error) {
	cmd := exec.Command(params[0], params[1:]...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("execute command error: %v, stderr: %s", err, stderr.String())
	}
	return out.String(), nil
}

func (p *Proxy) extractSignature(output string) (string, error) {
	re := regexp.MustCompile(`got signature\s+([A-Za-z0-9+/=]+)`)
	match := re.FindStringSubmatch(output)
	if len(match) < 2 {
		return "", errors.New("failed to parse signature")
	}
	return match[1], nil
}

func (p *Proxy) exportPub(validatorIDBase64 string) (string, error) {
	validatorKey := utils.MustBase64ToHexStr(validatorIDBase64, 0)
	params := p.buildCmd("exportpub", validatorKey)
	output, err := p.runCmd(params)
	if err != nil {
		return "", err
	}
	base64Result, err := p.extractPublicKey(output)
	if err != nil {
		return "", err
	}
	return utils.MustBase64ToHexStr(base64Result, 32), nil
}

func (p *Proxy) extractPublicKey(output string) (string, error) {
	re := regexp.MustCompile(`got public key:\s*([A-Za-z0-9+/=]+)`)
	match := re.FindStringSubmatch(output)
	if len(match) < 2 {
		return "", errors.New("failed to parse public key")
	}
	return match[1], nil
}

func (p *Proxy) extractCfg(output string) (*ValidatorEngineConfig, error) {
	re := regexp.MustCompile(`-+[\r\n]+([\s\S]*?)[\r\n]+-+`)
	match := re.FindStringSubmatch(output)
	if len(match) < 2 {
		return nil, errors.New("failed to parse JSON")
	}

	var config ValidatorEngineConfig
	if err := json.Unmarshal([]byte(strings.TrimSpace(match[1])), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return &config, nil
}

// private async runCmd(params: ValidatorEngineCommand) {
// 	const command = this._md
// 		`${this.validatorEngineConsolePath}/validator-engine-console`,
// 		params,
// 	);

// 	try {
// 		return await this.executeCommand(command);
// 	} catch (error) {
// 		throw new Error(`Error run cmd ${error}`);
// 	}
// }

// public async getValidatorConfig(): Promise<ValidatorEngineConfig> {
// 	const params: ValidatorEngineCommand = {
// 		...this.baseCommandParams,
// 		c: `getconfig`,
// 	};

// 	const result = await this.runCmd(params);
// 	return this.extractCfg(result!);
// }
