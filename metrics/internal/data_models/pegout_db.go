package data_models

import (
	"encoding/json"
	"errors"
	"fmt"

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
	Id               uint64
	Addr             *PegoutTonAddr
	Status           PegoutStatus
	BitcoinTxRaw     []byte
	BitcoinTxId      []byte
	BitcoinBlockHash []byte
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
	var pegout Pegout
	err := json.Unmarshal(jsonData, &pegout)
	if err != nil {
		return nil, err
	}

	return &pegout, nil
}

func DeserializePegoutsDB(jsonData []byte) ([]*Pegout, error) {
	var pegouts []*Pegout
	err := json.Unmarshal(jsonData, &pegouts)
	if err != nil {
		return nil, err
	}

	return pegouts, nil
}
