package ton

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func BuildExtMsg(
	unsignedMsgBody *cell.Cell,
	dstAddr *address.Address,
	signer signer.Signer,
) (*tlb.ExternalMessage, error) {
	signature := make([]byte, 64)
	if signer != nil {
		signature = signer.SignCell(unsignedMsgBody)
	}

	msgBody := cell.BeginCell().
		MustStoreSlice(signature, uint(len(signature)*8)).
		MustStoreBuilder(unsignedMsgBody.BeginParse().ToBuilder()).
		EndCell()

	msg := &tlb.ExternalMessage{
		DstAddr: dstAddr,
		Body:    msgBody,
	}

	return msg, nil
}
