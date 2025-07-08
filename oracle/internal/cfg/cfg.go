package cfg

type Cfg struct {
	TonConfigPathOrURL              string `env:"COMMON_TON_CONFIG,required"`
	CoordinatorContractAddr         string `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
	StandaloneMode                  bool   `env:"ORACLE_STANDALONE_MODE,required"`
	Pubkey                          string `env:"ORACLE_PUBKEY"`
	Secret                          string `env:"ORACLE_SECRET"`
	ValidatorEngineConsolePath      string `env:"ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH"`
	ServerPublicKeyPath             string `env:"ORACLE_SERVER_PUBLIC_KEY_PATH"`
	ClientPrivateKeyPath            string `env:"ORACLE_CLIENT_PRIVATE_KEY_PATH"`
	ValidatorServerAddr             string `env:"ORACLE_VALIDATOR_SERVER_ADDR"`
	KeystorePath                    string `env:"ORACLE_KEYSTORE_PATH"`
	FetchPeriod                     string `env:"ORACLE_DKG_FETCH_PERIOD"`
	SendStartDKGPeriod              string `env:"ORACLE_SEND_START_DKG_PERIOD"`
	ExecuteSignPeriod               string `env:"ORACLE_EXECUTE_SIGN_PERIOD"`
	ApiCallTimeout                  string `env:"API_CALL_TIMEOUT"`
	LogLevel                        string `env:"LOG_LEVEL"`
	LogFile                         string `env:"LOG_FILE"`
	LogMaxSize                      string `env:"LOG_FILE_MAX_SIZE"`
	LogMaxBackups                   string `env:"LOG_FILE_MAX_BACKUPS"`
	LogMaxBackupAge                 string `env:"LOG_FILE_MAX_BACKUP_AGE"`
	TestSkipR1                      bool   `env:"TEST_SKIP_R1"`
	TestSkipR2                      bool   `env:"TEST_SKIP_R2"`
	TestSkipR3                      bool   `env:"TEST_SKIP_R3"`
	TestSignSkipR1                  bool   `env:"TEST_SKIP_SIGN_R1"`
	TestSignSkipR2                  bool   `env:"TEST_SKIP_SIGN_R2"`
	TestSignSkipR3                  bool   `env:"TEST_SKIP_SIGN_R3"`
	TestBadR1Pkg                    bool   `env:"TEST_BAD_R1_PKG"`
	TestBadR1PkgRandomVersion       bool   `env:"TEST_BAD_R1_PKG_RANDOM_VERSION"`
	TestBadR2Pkg                    bool   `env:"TEST_BAD_R2_PKG"`
	TestBadR2Serialized             bool   `env:"TEST_BAD_R2_SERIALIZED"`
	TestBadR3Pkg                    bool   `env:"TEST_BAD_R3_PKG"`
	TestInvalidSigners              bool   `env:"TEST_INVALID_SIGNERS"`
	TestSignBadNonces               bool   `env:"TEST_SIGN_BAD_NONCES"`
	TestSignBadCommitments          bool   `env:"TEST_SIGN_BAD_COMMITMENTS"`
	TestSignBadShares               bool   `env:"TEST_SIGN_BAD_SHARES"`
	TestSignBadAggregatedSignatures bool   `env:"TEST_SIGN_BAD_AGGREGATED_SIGNATURES"`
}
