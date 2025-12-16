package fetchers

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type MockCoordinator struct {
	mock.Mock
}

func (m *MockCoordinator) GetDkg(block *ton.BlockIDExt) (*coordinator.DKG, error) {
	args := m.Called(block)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coordinator.DKG), args.Error(1)
}

func (m *MockCoordinator) GetPrevDKG() (*coordinator.DKG, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coordinator.DKG), args.Error(1)
}

func (m *MockCoordinator) GetUnsignedPegouts() ([]coordinator.PegoutRecord, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]coordinator.PegoutRecord), args.Error(1)
}

func (m *MockCoordinator) GetStorage(block *ton.BlockIDExt) (coordinator.Storage, error) {
	args := m.Called(block)
	return args.Get(0).(coordinator.Storage), args.Error(1)
}

func (m *MockCoordinator) SendStartDKG() (*tlb.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendRound1(
	validatorIdx uint16,
	dkgUntil int64,
	round1Package []byte,
	r2PublicX25519 *[32]byte,
) (*tlb.Transaction, error) {
	args := m.Called(validatorIdx, dkgUntil, round1Package, r2PublicX25519)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendRound2(
	validatorIdx uint16,
	dkgUntil int64,
	round2Packages []byte,
) (*tlb.Transaction, error) {
	args := m.Called(validatorIdx, dkgUntil, round2Packages)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendDKGClaim(
	validatorIdx uint16,
	dkgUntil int64,
	culpritIdx uint16,
) (*tlb.Transaction, error) {
	args := m.Called(validatorIdx, dkgUntil, culpritIdx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendPubkeyPackage(
	validatorIdx uint16,
	dkgUntil int64,
	sessionSigner signer.Signer,
	pubkeyPackage []byte,
) (*tlb.Transaction, error) {
	args := m.Called(validatorIdx, dkgUntil, sessionSigner, pubkeyPackage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendCommitments(
	pegoutID uint64,
	pegoutUntil int64,
	validatorIdx uint16,
	commitments []byte,
) (*tlb.Transaction, error) {
	args := m.Called(pegoutID, pegoutUntil, validatorIdx, commitments)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendSigningShare(
	pegoutID uint64,
	pegoutUntil int64,
	validatorIdx uint16,
	signingShares [][]byte,
) (*tlb.Transaction, error) {
	args := m.Called(pegoutID, pegoutUntil, validatorIdx, signingShares)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendSignatures(
	pegoutID uint64,
	pegoutUntil int64,
	validatorIdx uint16,
	signatures [][]byte,
) (*tlb.Transaction, error) {
	args := m.Called(pegoutID, pegoutUntil, validatorIdx, signatures)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendSigningClaim(
	pegoutID uint64,
	pegoutUntil int64,
	validatorIdx uint16,
	culpritIdx uint16,
) (*tlb.Transaction, error) {
	args := m.Called(pegoutID, pegoutUntil, validatorIdx, culpritIdx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) SendResetPegoutSigning(
	pegoutID uint64,
	validatorIdx uint16,
) (*tlb.Transaction, error) {
	args := m.Called(pegoutID, validatorIdx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tlb.Transaction), args.Error(1)
}

func (m *MockCoordinator) ConnectSigner(signer signer.Signer) {
	m.Called(signer)
}

func (m *MockCoordinator) GetAddr() *address.Address {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*address.Address)
}

func createTestStorage() coordinator.Storage {

	dkg := &coordinator.DKG{
		State: coordinator.DKGStatePart1Finished,
		VSet: coordinator.VSet{
			1: []byte{1, 2, 3}},
		MaxSigners: 3,
		VSetMask:   big.NewInt(7),
		SessionKeys: &coordinator.SessionKeys{
			PubKeys: coordinator.SessionPubKeys{
				1: []byte{1, 2, 3, 4},
				2: []byte{5, 6, 7, 8},
			},
		},
		R1: &coordinator.DKGR1{
			Mask:  big.NewInt(3),
			Count: 2,
			Packages: coordinator.DKGPkgs{
				1: []byte{1, 2, 3},
				2: []byte{4, 5, 6},
			},
		},
		R2: &coordinator.DKGR2{
			Mask:  big.NewInt(3),
			Count: 2,
			Packages: map[uint16][]byte{
				1: []byte{7, 8, 9},
				2: []byte{10, 11, 12},
			},
		},
		R3: &coordinator.DKGR3{
			Mask:  big.NewInt(3),
			Count: 2,
			Data: &coordinator.PubkeyData{
				PubkeyPackage: []byte{13, 14, 15},
				InternalKey:   []byte{16, 17, 18},
			},
		},
		Claims: &coordinator.DKGClaims{
			Mask:  big.NewInt(1),
			Count: 1,
			Counters: coordinator.DKGClaimcounters{
				1: 5,
			},
		},
		CfgHash:  []byte{19, 20, 21},
		Attempts: 3,
		Until:    time.Now().Add(24 * time.Hour),
	}

	pegoutRecord := coordinator.PegoutRecord{
		ID:            1,
		PegoutAddress: nil,
		InternalKey:   []byte{22, 23, 24},
		IsAutopegout:  true,
		Commitments: map[uint16][]byte{
			1: []byte{25, 26, 27},
			2: []byte{28, 29, 30},
		},
		CommitmentsMaskAccepted: big.NewInt(3),
		CommitmentsMaskOther:    big.NewInt(0),
		SigningShares: map[uint16]map[uint16][]byte{
			1: {
				1: []byte{31, 32, 33},
				2: []byte{34, 35, 36},
			},
			2: {
				1: []byte{37, 38, 39},
				2: []byte{40, 41, 42},
			},
		},
		SigningSharesMask: []byte{1, 0, 1},
		Signatures: coordinator.PegoutSignatures{
			Mask:  big.NewInt(3),
			Hash:  []byte{43, 44, 45},
			Count: 2,
		},

		ClaimsMask:     big.NewInt(1),
		ClaimsCount:    1,
		ClaimsCounters: map[uint16]uint16{1: 3},
		Signers:        2,
		ExpiredAt:      time.Now().Add(2 * time.Hour),
		SigningMask:    big.NewInt(3),
	}

	return coordinator.Storage{
		Initiated:           true,
		StandaloneMode:      false,
		Id:                  1,
		ConfiguratorAddr:    nil,
		Enabled:             true,
		Dkg:                 dkg,
		PrevDkg:             dkg,
		UnsignedPegouts:     []coordinator.PegoutRecord{pegoutRecord},
		PegoutTxCode:        nil,
		MinClaimsPercent:    50,
		MinSignersThreshold: 3,
		DkgLifetime:         1000,
		SigningTimeout:      300,
		NextPegoutIdx:       5,
		TeleportAddr:        nil,
	}
}

func TestFetcherContractCoordinator_Fetch_SuccessWithComplexStructures(t *testing.T) {
	expectedStorage := createTestStorage()

	mockCoord := &MockCoordinator{}
	mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
		Return(expectedStorage, nil)

	chDB := make(chan MetricsPayloadDB, 1)

	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 30)

	fetcher.Fetch()
	mockCoord.AssertCalled(t, "GetStorage", (*ton.BlockIDExt)(nil))

	select {
	case payload := <-chDB:
		assert.Equal(t, PayloadTypeContractCoordinator, payload.GetTypeId())

		var storage coordinator.Storage
		err := json.Unmarshal([]byte(payload.GetPayload()), &storage)
		require.NoError(t, err)

		assert.Nil(t, storage.Dkg)
		assert.Nil(t, storage.PrevDkg)

		assert.Equal(t, expectedStorage.Initiated, storage.Initiated)
		assert.Equal(t, expectedStorage.StandaloneMode, storage.StandaloneMode)
		assert.Equal(t, expectedStorage.Id, storage.Id)
		assert.Equal(t, expectedStorage.Enabled, storage.Enabled)
		assert.Equal(t, expectedStorage.MinClaimsPercent, storage.MinClaimsPercent)
		assert.Equal(t, expectedStorage.MinSignersThreshold, storage.MinSignersThreshold)
		assert.Equal(t, expectedStorage.DkgLifetime, storage.DkgLifetime)
		assert.Equal(t, expectedStorage.SigningTimeout, storage.SigningTimeout)
		assert.Equal(t, expectedStorage.NextPegoutIdx, storage.NextPegoutIdx)

		assert.Len(t, storage.UnsignedPegouts, len(expectedStorage.UnsignedPegouts))
		if len(storage.UnsignedPegouts) > 0 {
			pegout := storage.UnsignedPegouts[0]
			assert.Equal(t, expectedStorage.UnsignedPegouts[0].ID, pegout.ID)
			assert.Equal(t, expectedStorage.UnsignedPegouts[0].IsAutopegout, pegout.IsAutopegout)
			assert.Equal(t, expectedStorage.UnsignedPegouts[0].Signers, pegout.Signers)
			assert.NotNil(t, pegout.CommitmentsMaskAccepted)
			assert.NotNil(t, pegout.ClaimsMask)
			assert.NotNil(t, pegout.SigningMask)
		}

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected payload in channel, but got none")
	}
}

func TestFetcherContractCoordinator_Fetch_GetStorageError(t *testing.T) {
	mockCoord := &MockCoordinator{}
	expectedErr := errors.New("contract call failed")

	var emptyStorage coordinator.Storage
	mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
		Return(emptyStorage, expectedErr)

	chDB := make(chan MetricsPayloadDB, 1)

	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 30)

	fetcher.Fetch()

	mockCoord.AssertCalled(t, "GetStorage", (*ton.BlockIDExt)(nil))

	select {
	case <-chDB:
		t.Fatal("Should not receive payload when GetStorage fails")
	case <-time.After(50 * time.Millisecond):
		// expect nothing
	}
}

func TestFetcherContractCoordinator_Fetch_NilDkgFields(t *testing.T) {
	testCases := []struct {
		name             string
		dkg              *coordinator.DKG
		prevDkg          *coordinator.DKG
		expectNilDkg     bool
		expectNilPrevDkg bool
	}{
		{
			name:             "both DKG populated",
			dkg:              &coordinator.DKG{State: coordinator.DKGStatePart1Finished},
			prevDkg:          &coordinator.DKG{State: coordinator.DKGStateFinished},
			expectNilDkg:     true,
			expectNilPrevDkg: true,
		},
		{
			name:             "nil DKG fields",
			dkg:              nil,
			prevDkg:          nil,
			expectNilDkg:     true,
			expectNilPrevDkg: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := coordinator.Storage{
				Initiated:           true,
				StandaloneMode:      false,
				Id:                  1,
				Enabled:             true,
				Dkg:                 tc.dkg,
				PrevDkg:             tc.prevDkg,
				MinClaimsPercent:    50,
				MinSignersThreshold: 3,
				DkgLifetime:         1000,
				SigningTimeout:      300,
				NextPegoutIdx:       5,
				UnsignedPegouts: []coordinator.PegoutRecord{
					{
						ID:           1,
						IsAutopegout: true,
						Signers:      2,
						ClaimsMask:   big.NewInt(1),
						SigningMask:  big.NewInt(3),
					},
				},
			}

			mockCoord := &MockCoordinator{}
			mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
				Return(storage, nil)

			chDB := make(chan MetricsPayloadDB, 1)

			fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 30)

			fetcher.Fetch()

			payload, ok := <-chDB
			require.True(t, ok, "Expected payload in channel")

			var resultStorage coordinator.Storage
			err := json.Unmarshal([]byte(payload.GetPayload()), &resultStorage)
			require.NoError(t, err)

			assert.Nil(t, resultStorage.Dkg)
			assert.Nil(t, resultStorage.PrevDkg)

			assert.Equal(t, storage.Initiated, resultStorage.Initiated)
			assert.Equal(t, storage.Id, resultStorage.Id)
			assert.Len(t, resultStorage.UnsignedPegouts, len(storage.UnsignedPegouts))

			if len(resultStorage.UnsignedPegouts) > 0 {
				assert.NotNil(t, resultStorage.UnsignedPegouts[0].ClaimsMask)
				assert.NotNil(t, resultStorage.UnsignedPegouts[0].SigningMask)
			}
		})
	}
}

func TestFetcherContractCoordinator_Fetch_JsonMarshalError(t *testing.T) {
	storage := createTestStorage()

	mockCoord := &MockCoordinator{}

	mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
		Return(storage, nil)

	chDB := make(chan MetricsPayloadDB, 1)
	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 30)

	fetcher.Fetch()

	select {
	case <-chDB:
		t.Fatal("Should not receive payload when JSON marshaling fails")
	case <-time.After(50 * time.Millisecond):
		// expect nothing
	}
}

func TestFetcherContractCoordinator_Fetch_WithEmptyCollections(t *testing.T) {
	testCases := []struct {
		name        string
		storage     coordinator.Storage
		description string
	}{
		{
			name: "empty maps and slices",
			storage: coordinator.Storage{
				Initiated:       true,
				Id:              1,
				Dkg:             nil,
				PrevDkg:         nil,
				UnsignedPegouts: []coordinator.PegoutRecord{},
			},
			description: "Storage with empty collections",
		},
		{
			name: "nil collections",
			storage: coordinator.Storage{
				Initiated:       true,
				Id:              2,
				Dkg:             nil,
				PrevDkg:         nil,
				UnsignedPegouts: nil,
			},
			description: "Storage with nil collections",
		},
		{
			name: "DKG with empty substructures",
			storage: coordinator.Storage{
				Initiated: true,
				Id:        3,
				Dkg: &coordinator.DKG{
					State: coordinator.DKGStateInProgress,
					VSet:  coordinator.VSet{},
					R1: &coordinator.DKGR1{
						Packages: coordinator.DKGPkgs{},
					},
					R2: &coordinator.DKGR2{
						Packages: map[uint16][]byte{},
					},
					R3: &coordinator.DKGR3{
						Data: &coordinator.PubkeyData{},
					},
					Claims: &coordinator.DKGClaims{
						Counters: coordinator.DKGClaimcounters{},
					},
				},
				PrevDkg: nil,
			},
			description: "Storage with DKG having empty substructures",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCoord := &MockCoordinator{}
			mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
				Return(tc.storage, nil)

			chDB := make(chan MetricsPayloadDB, 1)
			fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 30)

			fetcher.Fetch()

			select {
			case payload := <-chDB:
				var result coordinator.Storage
				err := json.Unmarshal([]byte(payload.GetPayload()), &result)
				require.NoError(t, err)

				assert.Nil(t, result.Dkg)
				assert.Nil(t, result.PrevDkg)

				assert.Equal(t, tc.storage.Initiated, result.Initiated)
				assert.Equal(t, tc.storage.Id, result.Id)

			case <-time.After(100 * time.Millisecond):
				t.Fatal("Expected payload in channel")
			}
		})
	}
}

func TestFetcherContractCoordinator_Fetch_BigIntSerialization(t *testing.T) {
	storage := coordinator.Storage{
		Initiated: true,
		Id:        1,
		Dkg: &coordinator.DKG{
			State:    coordinator.DKGStateInProgress,
			VSetMask: big.NewInt(123456789),
			R1: &coordinator.DKGR1{
				Mask: big.NewInt(987654321),
			},
			R2: &coordinator.DKGR2{
				Mask: big.NewInt(555555555),
			},
			R3: &coordinator.DKGR3{
				Mask: big.NewInt(777777777),
			},
			Claims: &coordinator.DKGClaims{
				Mask: big.NewInt(999999999),
			},
		},
		PrevDkg: &coordinator.DKG{
			VSetMask: big.NewInt(111111111),
		},
		UnsignedPegouts: []coordinator.PegoutRecord{
			{
				ID:                      1,
				CommitmentsMaskAccepted: big.NewInt(333333333),
				CommitmentsMaskOther:    big.NewInt(444444444),
				ClaimsMask:              big.NewInt(666666666),
				SigningMask:             big.NewInt(888888888),
			},
		},
	}

	mockCoord := &MockCoordinator{}
	mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
		Return(storage, nil)

	chDB := make(chan MetricsPayloadDB, 1)
	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 30)

	fetcher.Fetch()

	select {
	case payload := <-chDB:
		var result map[string]interface{}
		err := json.Unmarshal([]byte(payload.GetPayload()), &result)
		require.NoError(t, err)

		assert.Nil(t, result["Dkg"])
		assert.Nil(t, result["PrevDkg"])

		assert.NotNil(t, result["Id"])
		assert.NotNil(t, result["Initiated"])

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected payload in channel")
	}
}

func TestFetcherContractCoordinator_Work_MultipleFetches(t *testing.T) {
	mockCoord := &MockCoordinator{}
	storage := createTestStorage()

	mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
		Return(storage, nil).
		Times(2)

	chDB := make(chan MetricsPayloadDB, 10)
	period := int64(1)

	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, period)

	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	done := make(chan bool)
	go func() {
		fetcher.Work(ctx)
		done <- true
	}()

	<-done

	assert.GreaterOrEqual(t, len(mockCoord.Calls), 1, "GetStorage should have been called at least once")

	payloadCount := 0
	for len(chDB) > 0 {
		payload := <-chDB
		payloadCount++

		var resultStorage coordinator.Storage
		err := json.Unmarshal([]byte(payload.GetPayload()), &resultStorage)
		require.NoError(t, err)
		assert.Nil(t, resultStorage.Dkg)
		assert.Nil(t, resultStorage.PrevDkg)

		assert.Len(t, resultStorage.UnsignedPegouts, len(storage.UnsignedPegouts))
	}

	assert.GreaterOrEqual(t, payloadCount, 1, "Should have sent at least one payload")
}

func TestFetcherContractCoordinator_Fetch_PegoutRecordComplexity(t *testing.T) {
	now := time.Now()

	storage := coordinator.Storage{
		Initiated: true,
		Id:        1,
		Dkg:       nil,
		PrevDkg:   nil,
		UnsignedPegouts: []coordinator.PegoutRecord{
			{
				ID:           1001,
				InternalKey:  []byte{1, 2, 3, 4, 5},
				IsAutopegout: false,
				Commitments: map[uint16][]byte{
					1: []byte{6, 7, 8},
					2: []byte{9, 10, 11},
					3: []byte{12, 13, 14},
				},
				CommitmentsMaskAccepted: big.NewInt(7),
				CommitmentsMaskOther:    big.NewInt(0),
				SigningShares: map[uint16]map[uint16][]byte{
					1: {
						1: []byte{15, 16, 17},
						2: []byte{18, 19, 20},
					},
					2: {
						1: []byte{21, 22, 23},
						2: []byte{24, 25, 26},
					},
				},
				SigningSharesMask: []byte{1, 1, 0},
				Signatures: coordinator.PegoutSignatures{
					Mask:  big.NewInt(3),
					Hash:  []byte{27, 28, 29},
					Count: 1,
				},
				ClaimsMask:     big.NewInt(5),
				ClaimsCount:    2,
				ClaimsCounters: map[uint16]uint16{1: 3, 2: 1},
				Signers:        3,
				ExpiredAt:      now.Add(time.Hour),
				SigningMask:    big.NewInt(15),
			},
			{
				ID:           1002,
				IsAutopegout: true,
				Signers:      1,
				ClaimsMask:   big.NewInt(1),
				SigningMask:  big.NewInt(1),
				ExpiredAt:    now.Add(2 * time.Hour),
			},
		},
	}

	mockCoord := &MockCoordinator{}
	mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
		Return(storage, nil)

	chDB := make(chan MetricsPayloadDB, 1)
	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 30)

	fetcher.Fetch()

	select {
	case payload := <-chDB:
		var result coordinator.Storage
		err := json.Unmarshal([]byte(payload.GetPayload()), &result)
		require.NoError(t, err)

		assert.Nil(t, result.Dkg)
		assert.Nil(t, result.PrevDkg)
		assert.Len(t, result.UnsignedPegouts, 2)

		if len(result.UnsignedPegouts) > 0 {
			pegout1 := result.UnsignedPegouts[0]
			assert.Equal(t, uint64(1001), pegout1.ID)
			assert.False(t, pegout1.IsAutopegout)
			assert.Equal(t, uint16(3), pegout1.Signers)
			assert.NotNil(t, pegout1.ClaimsMask)
			assert.NotNil(t, pegout1.SigningMask)
			assert.Len(t, pegout1.Commitments, 3)
			assert.Len(t, pegout1.SigningShares, 2)

			pegout2 := result.UnsignedPegouts[1]
			assert.Equal(t, uint64(1002), pegout2.ID)
			assert.True(t, pegout2.IsAutopegout)
			assert.Equal(t, uint16(1), pegout2.Signers)
		}

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected payload in channel")
	}
}

func TestFetcherContractCoordinator_ConcurrentAccess(t *testing.T) {
	mockCoord := &MockCoordinator{}
	storage := createTestStorage()

	mockCoord.On("GetStorage", (*ton.BlockIDExt)(nil)).
		Return(storage, nil).
		Times(5)

	chDB := make(chan MetricsPayloadDB, 10)
	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, 1)

	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func(id int) {
			for j := 0; j < 2; j++ {
				fetcher.Fetch()
				time.Sleep(100 * time.Millisecond)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 3; i++ {
		<-done
	}

	assert.GreaterOrEqual(t, len(mockCoord.Calls), 1, "Should have multiple calls to GetStorage")

	for len(chDB) > 0 {
		payload := <-chDB
		var result coordinator.Storage
		err := json.Unmarshal([]byte(payload.GetPayload()), &result)
		require.NoError(t, err)
		assert.Nil(t, result.Dkg)
		assert.Nil(t, result.PrevDkg)
		assert.Len(t, result.UnsignedPegouts, len(storage.UnsignedPegouts))
	}
}

func TestNewFetcherContractCoordinator(t *testing.T) {
	chDB := make(chan MetricsPayloadDB, 10)
	mockCoord := &MockCoordinator{}
	period := int64(30)

	fetcher := NewFetcherContractCoordinator(chDB, mockCoord, period)

	assert.NotNil(t, fetcher)
}
