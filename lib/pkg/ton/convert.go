package ton

import (
	"encoding/hex"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
)

func AddrToRawString(addr *address.Address) string {
	return fmt.Sprintf("%d:%s", addr.Workchain(), hex.EncodeToString(addr.Data()))
}
