package dkg

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
)

type Fetcher struct {
	coordinatorContract *coordinatorcontract.CoordinatorContract
}

func NewFetcher(contract *coordinatorcontract.CoordinatorContract) *Fetcher {
	return &Fetcher{
		coordinatorContract: contract,
	}
}

func (f *Fetcher) Fetch() (*coordinatorcontract.DKG, error) {
	return f.coordinatorContract.GetDkg(nil)
}
