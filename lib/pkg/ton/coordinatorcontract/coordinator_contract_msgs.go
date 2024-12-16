package coordinatorcontract

import (
	"log"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const DefaultDGKTTL = time.Minute * 5

func (c *CoordinatorContract) SendStartDKG(ttl int64) (*tlb.Transaction, error) {
	if ttl == 0 {
		ttl = int64(DefaultDGKTTL.Seconds())
	}

	unsignedMsgBody := cell.BeginCell().
		MustStoreUInt(OpCodeStartDKG, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		EndCell()

	log.Printf("unix: %d", uint64(time.Now().Unix()+ttl))

	log.Printf("addr: %+v", c.Addr)

	msg, err := ton.BuildExtMsg(unsignedMsgBody, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}

	celll, err := tlb.ToCell(msg)
	if err != nil {
		return nil, err
	}

	log.Printf("c: %+v", celll.BeginParse().BitsLeft())

	log.Printf("msg: %+v", msg.Payload().BeginParse().BitsLeft())
	log.Printf("msg: %+v", msg.Payload().BeginParse().RefsNum())
	log.Printf("msg: %+v", msg.Body.BeginParse().BitsLeft())
	log.Printf("msg: %+v", msg.Body.BeginParse().RefsNum())

	tx, block, inMsgHash, err := c.tonClient.API.SendExternalMessageWaitTransaction(c.ctx, msg)

	log.Printf("tx: %+v", tx)
	log.Printf("block: %+v", block)
	log.Printf("inMsgHash: %+v", inMsgHash)

	return tx, err
}
