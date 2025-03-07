package dkg

import "context"

type CommitResult struct {
	Nonce       []byte
	Commitments []byte
}

type DkgService interface {
	Commit(ctx context.Context, internalKey []byte) (*CommitResult, error)
	Sign(ctx context.Context, internalKey []byte, signPkg []byte, nonce []byte, merkleRoot []byte) ([]byte, error)
}
