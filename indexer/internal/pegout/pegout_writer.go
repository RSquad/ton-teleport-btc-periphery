package pegout

import (
	"context"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
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
) (*ent.Pegout, error) {
	pegoutContractAddr, err := pegoutcontract.AddrFromStateInit(
		&pegoutcontract.StateInit{
			Code: ew.pegoutContractCode,
			InitData: &pegoutcontract.InitData{
				ID:                   uint32(event.GetID()),
				Amount:               event.GetAmount(),
				BitcoinScript:        event.GetBitcoinScript(),
				TeleportContractAddr: ew.teleportAddr,
			},
		})
	if err != nil {
		return nil, err
	}

	pegout, err := ew.repo.Pegout.Create().
		SetExternalID(int64(event.GetID())).
		SetAddr(utils.AddrToRawString(pegoutContractAddr)).
		Save(ew.ctx)

	return pegout, err
}
