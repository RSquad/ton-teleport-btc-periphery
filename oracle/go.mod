module github.com/rsquad/ton-teleport-btc-periphery/oracle

go 1.23.1

replace github.com/rsquad/ton-teleport-btc-periphery/lib => ../lib

replace github.com/rsquad/ton-teleport-btc-periphery/frost => ../frost

require (
	github.com/rs/zerolog v1.33.0
	github.com/rsquad/ton-teleport-btc-periphery/frost v0.0.0-00010101000000-000000000000
	github.com/rsquad/ton-teleport-btc-periphery/lib v0.0.0-00010101000000-000000000000
	github.com/xssnick/tonutils-go v1.11.1
)

require (
	github.com/btcsuite/btcd v0.24.2 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.1.3 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.5 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd // indirect
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792 // indirect
	github.com/caarlos0/env/v11 v11.2.2 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20220328075252-7dd334e3daae // indirect
	github.com/sigurn/crc16 v0.0.0-20211026045750-20ab5afb07e3 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
