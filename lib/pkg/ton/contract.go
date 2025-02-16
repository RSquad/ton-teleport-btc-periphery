package ton

import "github.com/xssnick/tonutils-go/address"

type ContractInterface interface {
	GetAddr() *address.Address
}

type Contract struct {
	Addr *address.Address
}

func (c *Contract) GetAddr() *address.Address {
	return c.Addr
}
