package dkg

import "context"

type CommitRequest struct {
	internalKey []byte
}

type CommitResult struct {
	Nonce       []byte
	Commitments []byte
}

type SignRequest struct {
	internalKey []byte
	signPkg     []byte
	nonce       []byte
	merkleRoot  []byte
}

type SignResult struct{}

type DkgService interface {
	Commit(ctx context.Context, internalKey []byte) (*CommitResult, error)
	Sign(ctx context.Context, internalKey []byte, signPkg []byte, nonce []byte, merkleRoot []byte) ([]byte, error)
}

type Client struct {
	commitRequestCh chan *CommitRequest
	commitResultCh  chan *CommitResult
	signRequestCh   chan *SignRequest
	signResultCh    chan *SignResult
}
