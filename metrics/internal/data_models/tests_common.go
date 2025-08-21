package data_models

import (
	"encoding/json"
	"testing"
)

func MustUnmarshalJSONArray(t *testing.T, js string) []interface{} {
	t.Helper()
	var arr []interface{}
	if err := json.Unmarshal([]byte(js), &arr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return arr
}
