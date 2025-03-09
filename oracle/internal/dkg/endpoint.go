package dkg

type CommitRequest struct {
	internalKey []byte
}

type CommitResult struct {
	Nonce       []byte
	Commitments []byte
}

type SignRequest struct {
	internalKey []byte
	tapTweak    []byte
	pegoutAddr  string
	message     []byte
}

type SignResult struct {
	signingShare []byte
}

type Endpoint struct {
	CommitRequestCh chan *CommitRequest
	CommitResultCh  chan *CommitResult
	SignRequestCh   chan *SignRequest
	SignResultCh    chan *SignResult
}

func CreateEndpoint() *Endpoint {
	return &Endpoint{
		CommitRequestCh: make(chan *CommitRequest),
		CommitResultCh:  make(chan *CommitResult),
		SignRequestCh:   make(chan *SignRequest),
		SignResultCh:    make(chan *SignResult),
	}
}
