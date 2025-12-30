package pegout

import (
	"context"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type PegoutWriter struct {
	ctx                context.Context
	repo               *ent.Client
	teleportAddr       *address.Address
	pegoutContractCode *cell.Cell
}

func NewPegoutWriter(
	ctx context.Context,
	repo *ent.Client,
	teleportAddr *address.Address,
	pegoutContractCode *cell.Cell,
) *PegoutWriter {
	return &PegoutWriter{
		ctx:                ctx,
		repo:               repo,
		teleportAddr:       teleportAddr,
		pegoutContractCode: pegoutContractCode,
	}
}

func (ew *PegoutWriter) WriteFromEvent(
	event teleportcontract.EventWithPegoutInterface,
	tx *ent.Tx,
) (*ent.Pegout, error) {
	var pegoutCli *ent.PegoutClient = nil
	if tx != nil {
		pegoutCli = tx.Pegout
	} else {
		pegoutCli = ew.repo.Pegout
	}

	pegout, err := pegoutCli.Create().
		SetAddr(func() string {
			if event.GetPegoutAddr().IsAddrNone() {
				return ""
			}
			return utils.AddrToRawString(event.GetPegoutAddr())
		}()).
		Save(ew.ctx)

	if err != nil {
		logPegoutCreateError(err, event)
	}

	return pegout, err
}
