package data_models

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

func ToString(v interface{}) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	default:
		return "", fmt.Errorf("expected string, got %T", v)
	}
}

func ToBool(v interface{}) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		b, err := strconv.ParseBool(t)
		if err != nil {
			return false, fmt.Errorf("invalid bool string %q", t)
		}
		return b, nil
	default:
		return false, fmt.Errorf("expected bool, got %T", v)
	}
}

func ToUint64(v interface{}) (uint64, error) {
	switch t := v.(type) {
	case float64:
		if t < 0 || t > math.MaxUint64 {
			return 0, fmt.Errorf("float64 out of uint64 range: %v", t)
		}
		if math.Trunc(t) != t {
			return 0, fmt.Errorf("non-integer float64: %v", t)
		}
		return uint64(t), nil
	case int64:
		if t < 0 {
			return 0, fmt.Errorf("negative int64: %v", t)
		}
		return uint64(t), nil
	case int:
		if t < 0 {
			return 0, fmt.Errorf("negative int: %v", t)
		}
		return uint64(t), nil
	case uint64:
		return t, nil
	case string:
		u, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid uint64 string %q: %w", t, err)
		}
		return u, nil
	default:
		return 0, fmt.Errorf("expected number/string, got %T", v)
	}
}

func ToBigInt(v interface{}) (*big.Int, error) {
	switch t := v.(type) {
	case string:
		z, ok := new(big.Int).SetString(t, 10)
		if !ok {
			return nil, fmt.Errorf("invalid big.Int string %q", t)
		}
		return z, nil
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return nil, errors.New("invalid float value for big.Int")
		}
		return new(big.Int).SetUint64(uint64(t)), nil
	case int64:
		return big.NewInt(t), nil
	case uint64:
		return new(big.Int).SetUint64(t), nil
	case *big.Int:
		return new(big.Int).Set(t), nil
	default:
		return nil, fmt.Errorf("expected big-int-like, got %T", v)
	}
}

func MapStringToBytesByUint16Key(v interface{}) (map[uint16][]byte, error) {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for map[uint16][]byte, got %T", v)
	}
	out := make(map[uint16][]byte, len(obj))
	for k, vv := range obj {
		keyU, err := ParseUint16Key(k)
		if err != nil {
			return nil, err
		}
		s, err := ToString(vv)
		if err != nil {
			return nil, fmt.Errorf("value for key %q: %w", k, err)
		}
		bs, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("base64 decode for key %q: %w", k, err)
		}
		out[keyU] = bs
	}
	return out, nil
}

func MapStringToUint16ByUint16Key(v interface{}) (map[uint16]uint16, error) {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for map[uint16]uint16, got %T", v)
	}
	out := make(map[uint16]uint16, len(obj))
	for k, vv := range obj {
		keyU, err := ParseUint16Key(k)
		if err != nil {
			return nil, err
		}
		u64, err := ToUint64(vv)
		if err != nil {
			return nil, fmt.Errorf("value for key %q: %w", k, err)
		}
		out[keyU] = uint16(u64)
	}
	return out, nil
}

func Map2DStringToBytesByUint16Keys(v interface{}) (map[uint16]map[uint16][]byte, error) {
	level1, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for map[uint16]map[uint16][]byte, got %T", v)
	}
	out := make(map[uint16]map[uint16][]byte, len(level1))
	for k1, inner := range level1 {
		key1, err := ParseUint16Key(k1)
		if err != nil {
			return nil, err
		}
		if inner == nil {
			out[key1] = map[uint16][]byte{}
			continue
		}
		level2, ok := inner.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("inner value for %q is not object", k1)
		}
		m2 := make(map[uint16][]byte, len(level2))
		for k2, vv := range level2 {
			key2, err := ParseUint16Key(k2)
			if err != nil {
				return nil, err
			}
			s, err := ToString(vv)
			if err != nil {
				return nil, fmt.Errorf("value for %q.%q: %w", k1, k2, err)
			}
			bs, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, fmt.Errorf("base64 decode for %q.%q: %w", k1, k2, err)
			}
			m2[key2] = bs
		}
		out[key1] = m2
	}
	return out, nil
}

func ParseUint16Key(s string) (uint16, error) {
	u, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid uint16 key %q: %w", s, err)
	}
	return uint16(u), nil
}
