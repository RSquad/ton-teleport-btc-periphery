package data_models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/xssnick/tonutils-go/address"
)

type PegoutStatus int

const (
	PEGOUT_CONFIRMED PegoutStatus = iota
	PEGOUT_SIGNED
	PEGOUT_SIGNING
)

type PegoutTonAddr address.Address

type Pegout struct {
	Id               uint64         `json:"id"`
	Addr             *PegoutTonAddr `json:"addr"`
	Status           PegoutStatus   `json:"status"`
	BitcoinTxRaw     []byte         `json:"bitcoin_tx_raw"`
	BitcoinTxId      []byte         `json:"bitcoin_tx_id"`
	BitcoinBlockHash []byte         `json:"bitcoin_block_hash"`
}

type PegoutJSON struct {
	Id                  uint64         `json:"id"`
	Addr                *PegoutTonAddr `json:"addr"`
	Status              PegoutStatus   `json:"status"`
	BitcoinTxRawHex     string         `json:"bitcoin_tx_raw"`
	BitcoinTxIdHex      string         `json:"bitcoin_tx_id"`
	BitcoinBlockHashHex string         `json:"bitcoin_block_hash"`
}

func (s *PegoutTonAddr) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid data")
	}

	data = data[1 : len(data)-1]
	strData := string(data)
	addr, err := address.ParseRawAddr(strData)
	if err != nil {
		return err
	}

	*s = *(*PegoutTonAddr)(addr)

	return nil
}

const (
	statusConfirmed = "CONFIRMED"
	statusSigned    = "SIGNED"
	statusSigning   = "SIGNING"
)

var toString = map[PegoutStatus]string{
	PEGOUT_CONFIRMED: statusConfirmed,
	PEGOUT_SIGNED:    statusSigned,
	PEGOUT_SIGNING:   statusSigning,
}

var fromString = map[string]PegoutStatus{
	statusConfirmed: PEGOUT_CONFIRMED,
	statusSigned:    PEGOUT_SIGNED,
	statusSigning:   PEGOUT_SIGNING,
}

func (s PegoutStatus) String() string {
	if v, ok := toString[s]; ok {
		return v
	}
	return fmt.Sprintf("PegoutStatus(%d)", int(s))
}

func (s PegoutStatus) MarshalText() ([]byte, error) {
	if v, ok := toString[s]; ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("unknown PegoutStatus value: %d", int(s))
}

func (s *PegoutStatus) UnmarshalText(text []byte) error {
	if s == nil {
		return errors.New("nil receiver")
	}
	if v, ok := fromString[string(text)]; ok {
		*s = v
		return nil
	}
	return fmt.Errorf("invalid PegoutStatus: %q", string(text))
}

func (s *PegoutStatus) UnmarshalJSON(b []byte) error {
	var asString string
	if err := json.Unmarshal(b, &asString); err == nil {
		return s.UnmarshalText([]byte(asString))
	}

	var asNum int
	if err := json.Unmarshal(b, &asNum); err == nil {
		ps := PegoutStatus(asNum)
		if _, ok := toString[ps]; !ok {
			return fmt.Errorf("invalid PegoutStatus numeric: %d", asNum)
		}
		*s = ps
		return nil
	}
	return fmt.Errorf("PegoutStatus must be string or number, got: %s", string(b))
}

func (s PegoutStatus) MarshalJSON() ([]byte, error) {
	txt, err := s.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(txt))
}

func DeserializePegoutDB(jsonData []byte) (*Pegout, error) {
	var pegoutJson PegoutJSON
	err := json.Unmarshal(jsonData, &pegoutJson)
	if err != nil {
		return nil, fmt.Errorf("failed to call `DeserializePegoutDB`, json '%s': %w", string(jsonData), err)
	}

	return PegoutJsonToPegout(&pegoutJson)
}

func DeserializePegoutsDB(jsonData []byte) ([]*Pegout, error) {
	var pegoutsJson []*PegoutJSON
	err := json.Unmarshal(jsonData, &pegoutsJson)
	if err != nil {
		return nil, fmt.Errorf("failed to call `DeserializePegoutsDB`, json '%s': %w", string(jsonData), err)
	}

	pegouts := make([]*Pegout, len(pegoutsJson))

	for i, pegoutJson := range pegoutsJson {
		pegout, err := PegoutJsonToPegout(pegoutJson)
		if err != nil {
			return nil, err
		}

		pegouts[i] = pegout
	}

	return pegouts, nil
}

func PegoutJsonToPegout(pegoutJson *PegoutJSON) (*Pegout, error) {
	h := func(s string) ([]byte, error) {
		if s == "" {
			return nil, nil
		}
		return hex.DecodeString(strings.TrimPrefix(strings.ToLower(s), "0x"))
	}

	bitcoinTxRaw, err := h(pegoutJson.BitcoinTxRawHex)
	if err != nil {
		return nil, err
	}

	bitcoinTxId, err := h(pegoutJson.BitcoinTxIdHex)
	if err != nil {
		return nil, err
	}

	bitcoinBlockHash, err := h(pegoutJson.BitcoinBlockHashHex)
	if err != nil {
		return nil, err
	}

	var pegout Pegout

	pegout.Id = pegoutJson.Id
	pegout.Addr = pegoutJson.Addr
	pegout.Status = pegoutJson.Status
	pegout.BitcoinTxRaw = bitcoinTxRaw
	pegout.BitcoinTxId = bitcoinTxId
	pegout.BitcoinBlockHash = bitcoinBlockHash

	return &pegout, nil
}
