module github.com/rsquad/ton-teleport-btc-periphery/oracle

go 1.23.1

toolchain go1.23.3

replace github.com/rsquad/ton-teleport-btc-periphery/lib => ../lib

require (
	github.com/rsquad/ton-teleport-btc-periphery/lib v0.0.0-00010101000000-000000000000
	github.com/xssnick/tonutils-go v1.10.2
)

require (
	github.com/caarlos0/env/v11 v11.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gojuno/minimock/v3 v3.4.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20220328075252-7dd334e3daae // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sigurn/crc16 v0.0.0-20211026045750-20ab5afb07e3 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
)
