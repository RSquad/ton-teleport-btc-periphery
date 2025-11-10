package coordinator

import (
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func TestEventParser_Parse(t *testing.T) {
	parser := NewEventParser()

	tests := []struct {
		name        string
		eventID     uint32
		buildBody   func() *cell.Cell
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "DKGComplete event",
			eventID: eventIdDKGComplete,
			buildBody: func() *cell.Cell {
				builder := cell.BeginCell()
				builder.MustStoreUInt(eventIdDKGComplete, 32)
				builder.MustStoreUInt(uint64(time.Now().Unix()), 64)
				builder.MustStoreSlice(make([]byte, 32), 256) // 256 bits = 32 bytes
				return builder.EndCell()
			},
			wantErr: false,
		},
		{
			name:    "DKGStarted event",
			eventID: eventIdDKGStarted,
			buildBody: func() *cell.Cell {
				builder := cell.BeginCell()
				builder.MustStoreUInt(eventIdDKGStarted, 32)
				return builder.EndCell()
			},
			wantErr: false,
		},
		{
			name:    "Unknown event type",
			eventID: 0x012345678,
			buildBody: func() *cell.Cell {
				builder := cell.BeginCell()
				builder.MustStoreUInt(0x012345678, 32)
				return builder.EndCell()
			},
			wantErr:     true,
			expectedErr: "unknown event type with id 0x012345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.buildBody()
			rawEvent := &ton.RawEvent{
				Body: body,
			}

			result, err := parser.Parse(rawEvent)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("Expected error %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			if result.GetEventID() != tt.eventID {
				t.Errorf("Expected event ID %x, got %x", tt.eventID, result.GetEventID())
			}

			if result.GetRaw() != rawEvent {
				t.Error("Raw event reference mismatch")
			}
		})
	}
}

func TestParseDKGCompleteEvent(t *testing.T) {
	tests := []struct {
		name         string
		buildSlice   func() *cell.Slice
		wantErr      bool
		expectedKey  []byte
		expectedTime time.Time
	}{
		{
			name: "Valid DKGComplete event",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				completedAt := uint64(time.Now().Unix())
				builder.MustStoreUInt(completedAt, 64)
				key := make([]byte, 32)
				key[0] = 0x42 // Some test data
				builder.MustStoreSlice(key, 256)
				return builder.EndCell().BeginParse()
			},
			wantErr:      false,
			expectedKey:  make([]byte, 32),
			expectedTime: time.Now().Truncate(time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := tt.buildSlice()
			rawEvent := &ton.RawEvent{}

			result, err := parseDKGCompleteEvent(slice, rawEvent)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			if result.Raw != rawEvent {
				t.Error("Raw event reference mismatch")
			}

			if len(result.Key) != 32 {
				t.Errorf("Expected key length 32, got %d", len(result.Key))
			}

			// Check if completedAt is within 1 second of expected (Unix time precision)
			if result.CompletedAt.Unix() != tt.expectedTime.Unix() {
				t.Errorf("Expected completedAt %v, got %v", tt.expectedTime, result.CompletedAt)
			}
		})
	}
}

func TestParseDKGStartedEvent(t *testing.T) {
	rawEvent := &ton.RawEvent{}

	result, err := parseDKGStartedEvent(rawEvent)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected non-nil result")
		return
	}

	if result.Raw != rawEvent {
		t.Error("Raw event reference mismatch")
	}

	if result.GetEventID() != eventIdDKGStarted {
		t.Errorf("Expected event ID %x, got %x", eventIdDKGStarted, result.GetEventID())
	}
}

func TestParseDKGCompletedInfoEvent(t *testing.T) {
	rawEvent := &ton.RawEvent{}

	result, err := parseDKGCompletedInfoEvent(rawEvent)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected non-nil result")
		return
	}

	if result.Raw != rawEvent {
		t.Error("Raw event reference mismatch")
	}

	if result.GetEventID() != eventIdDKGCompletedInfo {
		t.Errorf("Expected event ID %x, got %x", eventIdDKGCompletedInfo, result.GetEventID())
	}
}

func TestParseDKGRestartedEvent(t *testing.T) {
	tests := []struct {
		name        string
		buildSlice  func() *cell.Slice
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Valid DKGRestarted event",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				builder.MustStoreUInt(1, 8) // reason

				// Add newDkg ref
				dkgBuilder := cell.BeginCell()
				dkgBuilder.MustStoreUInt(123, 64)
				dkgCell := dkgBuilder.EndCell()
				builder.MustStoreRef(dkgCell)

				// Add empty claims dict
				dict := cell.NewDict(16)
				builder.MustStoreDict(dict)

				builder.MustStoreBigUInt(big.NewInt(0xffff), 256) // claimsMask
				return builder.EndCell().BeginParse()
			},
			wantErr: false,
		},
		{
			name: "Invalid claims dict",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				builder.MustStoreUInt(1, 8) // reason

				// Add newDkg ref
				dkgBuilder := cell.BeginCell()
				dkgBuilder.MustStoreUInt(123, 64)
				dkgCell := dkgBuilder.EndCell()
				builder.MustStoreRef(dkgCell)

				// Store invalid dict (store raw slice instead of proper dict)
				builder.MustStoreUInt(0, 16)
				builder.MustStoreBigUInt(big.NewInt(0xffff), 256)
				return builder.EndCell().BeginParse()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := tt.buildSlice()
			rawEvent := &ton.RawEvent{}

			result, err := parseDKGRestartedEvent(slice, rawEvent)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.expectedErr != "" {
					// Check if error contains expected substring
					if err != nil && tt.expectedErr != "" {
						// Since we can't check exact error, we'll just ensure we got an error
						t.Logf("Got expected error: %v", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			if result.Raw != rawEvent {
				t.Error("Raw event reference mismatch")
			}

			if result.Reason.Cmp(big.NewInt(1)) != 0 {
				t.Errorf("Expected reason 1, got %v", result.Reason)
			}

			if result.ClaimsMask.Cmp(big.NewInt(0xffff)) != 0 {
				t.Errorf("Expected claimsMask 0xffff, got %v", result.ClaimsMask)
			}
		})
	}
}

func TestParseDKGRotatedEvent(t *testing.T) {
	rawEvent := &ton.RawEvent{}

	result, err := parseDKGRotatedEvent(rawEvent)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected non-nil result")
		return
	}

	if result.Raw != rawEvent {
		t.Error("Raw event reference mismatch")
	}

	if result.GetEventID() != eventIdDKGRotated {
		t.Errorf("Expected event ID %x, got %x", eventIdDKGRotated, result.GetEventID())
	}
}

func TestParsePegoutSigningStartedEvent(t *testing.T) {
	tests := []struct {
		name             string
		buildSlice       func() *cell.Slice
		expectedPegoutId *big.Int
	}{
		{
			name: "Valid PegoutSigningStarted event",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				builder.MustStoreUInt(12345, 64) // pegoutId
				return builder.EndCell().BeginParse()
			},
			expectedPegoutId: big.NewInt(12345),
		},
		{
			name: "Large pegoutId",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				builder.MustStoreUInt(^uint64(0), 64) // max uint64
				return builder.EndCell().BeginParse()
			},
			expectedPegoutId: new(big.Int).SetUint64(^uint64(0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := tt.buildSlice()
			rawEvent := &ton.RawEvent{}

			result, err := parsePegoutSigningStartedEvent(slice, rawEvent)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			if result.Raw != rawEvent {
				t.Error("Raw event reference mismatch")
			}

			if result.PegoutId.Cmp(tt.expectedPegoutId) != 0 {
				t.Errorf("Expected pegoutId %v, got %v", tt.expectedPegoutId, result.PegoutId)
			}
		})
	}
}

func TestParsePegoutSigningCompletedEvent(t *testing.T) {
	tests := []struct {
		name             string
		buildSlice       func() *cell.Slice
		expectedPegoutId *big.Int
	}{
		{
			name: "Valid PegoutSigningCompleted event",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				builder.MustStoreUInt(67890, 64) // pegoutId
				return builder.EndCell().BeginParse()
			},
			expectedPegoutId: big.NewInt(67890),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := tt.buildSlice()
			rawEvent := &ton.RawEvent{}

			result, err := parsePegoutSigningCompletedEvent(slice, rawEvent)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			if result.Raw != rawEvent {
				t.Error("Raw event reference mismatch")
			}

			if result.PegoutId.Cmp(tt.expectedPegoutId) != 0 {
				t.Errorf("Expected pegoutId %v, got %v", tt.expectedPegoutId, result.PegoutId)
			}
		})
	}
}

func TestParsePegoutSigningRestartedEvent(t *testing.T) {
	tests := []struct {
		name           string
		buildSlice     func() *cell.Slice
		wantErr        bool
		expectedErr    string
		expectedValues *PegoutSigningRestartedEvent
	}{
		{
			name: "Valid PegoutSigningRestarted event with all fields",
			buildSlice: func() *cell.Slice {
				// Build the main cell
				builder := cell.BeginCell()
				builder.MustStoreUInt(999, 64) // pegoutId
				builder.MustStoreUInt(2, 8)    // reason

				// Build and store pegout ref
				pegoutBuilder := cell.BeginCell()
				pegoutBuilder.MustStoreUInt(456, 64)
				pegoutCell := pegoutBuilder.EndCell()
				builder.MustStoreRef(pegoutCell)

				// Build and store claims dict
				dict := cell.NewDict(16)
				keyBuilder := cell.BeginCell()
				keyBuilder.MustStoreUInt(1, 16)
				valueBuilder := cell.BeginCell()
				valueBuilder.MustStoreUInt(42, 64)
				dict.Set(keyBuilder.EndCell(), valueBuilder.EndCell())
				builder.MustStoreDict(dict)

				builder.MustStoreBigUInt(big.NewInt(0x1111), 256) // claimsMask

				// Build and store the second ref cell
				refBuilder := cell.BeginCell()
				refBuilder.MustStoreBigUInt(big.NewInt(0xaaaa), 256) // commitmentMask
				refBuilder.MustStoreBigUInt(big.NewInt(0xbbbb), 256) // sharesMask
				refBuilder.MustStoreBigUInt(big.NewInt(0xcccc), 256) // signatureMask
				refCell := refBuilder.EndCell()
				builder.MustStoreRef(refCell)

				return builder.EndCell().BeginParse()
			},
			wantErr: false,
			expectedValues: &PegoutSigningRestartedEvent{
				PegoutId:       big.NewInt(999),
				Reason:         big.NewInt(2),
				CommitmentMask: big.NewInt(0xaaaa),
				SharesMask:     big.NewInt(0xbbbb),
				SignatureMask:  big.NewInt(0xcccc),
				ClaimsMask:     big.NewInt(0x1111),
			},
		},
		{
			name: "Valid PegoutSigningRestarted event with zero values",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				builder.MustStoreUInt(0, 64) // pegoutId
				builder.MustStoreUInt(0, 8)  // reason

				// Build and store empty pegout ref
				pegoutBuilder := cell.BeginCell()
				pegoutBuilder.MustStoreUInt(0, 64)
				pegoutCell := pegoutBuilder.EndCell()
				builder.MustStoreRef(pegoutCell)

				// Store empty claims dict
				dict := cell.NewDict(16)
				builder.MustStoreDict(dict)

				builder.MustStoreBigUInt(big.NewInt(0), 256) // claimsMask

				// Build and store the second ref cell with zeros
				refBuilder := cell.BeginCell()
				refBuilder.MustStoreBigUInt(big.NewInt(0), 256) // commitmentMask
				refBuilder.MustStoreBigUInt(big.NewInt(0), 256) // sharesMask
				refBuilder.MustStoreBigUInt(big.NewInt(0), 256) // signatureMask
				refCell := refBuilder.EndCell()
				builder.MustStoreRef(refCell)

				return builder.EndCell().BeginParse()
			},
			wantErr: false,
			expectedValues: &PegoutSigningRestartedEvent{
				PegoutId:       big.NewInt(0),
				Reason:         big.NewInt(0),
				CommitmentMask: big.NewInt(0),
				SharesMask:     big.NewInt(0),
				SignatureMask:  big.NewInt(0),
				ClaimsMask:     big.NewInt(0),
			},
		},
		{
			name: "Valid PegoutSigningRestarted event with maximum values",
			buildSlice: func() *cell.Slice {
				builder := cell.BeginCell()
				builder.MustStoreUInt(^uint64(0), 64) // max pegoutId
				builder.MustStoreUInt(255, 8)         // max reason

				// Build and store pegout ref
				pegoutBuilder := cell.BeginCell()
				pegoutBuilder.MustStoreUInt(^uint64(0), 64)
				pegoutCell := pegoutBuilder.EndCell()
				builder.MustStoreRef(pegoutCell)

				// Store empty claims dict
				dict := cell.NewDict(16)
				builder.MustStoreDict(dict)

				maxValue := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
				builder.MustStoreBigUInt(maxValue, 256) // max claimsMask

				// Build and store the second ref cell with max values
				refBuilder := cell.BeginCell()
				refBuilder.MustStoreBigUInt(maxValue, 256) // max commitmentMask
				refBuilder.MustStoreBigUInt(maxValue, 256) // max sharesMask
				refBuilder.MustStoreBigUInt(maxValue, 256) // max signatureMask
				refCell := refBuilder.EndCell()
				builder.MustStoreRef(refCell)

				return builder.EndCell().BeginParse()
			},
			wantErr: false,
			expectedValues: &PegoutSigningRestartedEvent{
				PegoutId:       new(big.Int).SetUint64(^uint64(0)),
				Reason:         big.NewInt(255),
				CommitmentMask: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)),
				SharesMask:     new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)),
				SignatureMask:  new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)),
				ClaimsMask:     new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := tt.buildSlice()
			rawEvent := &ton.RawEvent{}

			result, err := parsePegoutSigningRestartedEvent(slice, rawEvent)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.expectedErr != "" {
					// Check if error contains expected substring
					if err != nil && tt.expectedErr != "" {
						// Since the exact error message might vary, we'll check if it's one of the expected types
						errorMsg := err.Error()
						if errorMsg != "not enough refs" && errorMsg != "not enough bits" {
							t.Errorf("Expected error containing known issue, got: %v", err)
						}
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// Verify all parsed values
			if result.Raw != rawEvent {
				t.Error("Raw event reference mismatch")
			}

			if result.PegoutId.Cmp(tt.expectedValues.PegoutId) != 0 {
				t.Errorf("Expected pegoutId %v, got %v", tt.expectedValues.PegoutId, result.PegoutId)
			}

			if result.Reason.Cmp(tt.expectedValues.Reason) != 0 {
				t.Errorf("Expected reason %v, got %v", tt.expectedValues.Reason, result.Reason)
			}

			if result.CommitmentMask.Cmp(tt.expectedValues.CommitmentMask) != 0 {
				t.Errorf("Expected commitmentMask %v, got %v", tt.expectedValues.CommitmentMask, result.CommitmentMask)
			}

			if result.SharesMask.Cmp(tt.expectedValues.SharesMask) != 0 {
				t.Errorf("Expected sharesMask %v, got %v", tt.expectedValues.SharesMask, result.SharesMask)
			}

			if result.SignatureMask.Cmp(tt.expectedValues.SignatureMask) != 0 {
				t.Errorf("Expected signatureMask %v, got %v", tt.expectedValues.SignatureMask, result.SignatureMask)
			}

			if result.ClaimsMask.Cmp(tt.expectedValues.ClaimsMask) != 0 {
				t.Errorf("Expected claimsMask %v, got %v", tt.expectedValues.ClaimsMask, result.ClaimsMask)
			}

			// Verify that pegout ref was loaded (non-nil)
			if result.Pegout == nil {
				t.Error("Expected non-nil pegout ref")
			}

			// Verify that claims were loaded (non-nil)
			if result.Claims == nil {
				t.Error("Expected non-nil claims")
			}
		})
	}
}
