package teleportcontract

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3client/blockchain"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3models"
)

type TeleportContractLogListener struct {
	tonCenterV3Client *ton.TonCenterV3Client
	listenAddr        *address.Address
	offset            int64
	limit             int64
	onLogReceived     func(*cell.Cell)
}

func NewTeleportContractLogListener(
	tonCenterV3Client *ton.TonCenterV3Client,
	listenAddr *address.Address,
	onLogReceived func(*cell.Cell),
) (
	*TeleportContractLogListener,
	error,
) {
	return &TeleportContractLogListener{
		tonCenterV3Client,
		listenAddr,
		0,
		1000,
		onLogReceived,
	}, nil
}

func (c *TeleportContractLogListener) StartListen() {
	log.Println("[LogListener] listening started")
	c.listen()
}

func (c *TeleportContractLogListener) listen() {
	for {
		msgs, err := c.fetchMsgs()
		if err != nil {
			log.Println(fmt.Errorf("[LogListener] failed to fetch msgs %v", err))
		}

		c.processMsgs(msgs)

		c.offset += int64(len(msgs))

		if int64(len(msgs)) < c.limit {
			log.Println("[LogListener] all logs fetched, waiting for new logs")
			time.Sleep(3 * time.Second)
		}
	}
}

func (c *TeleportContractLogListener) fetchMsgs() ([]*toncenterv3models.Message, error) {
	params := blockchain.NewAPIV3GetMessagesParamsWithTimeout(30 * time.Second)
	src := c.listenAddr.String()
	params.SetSource(&src)
	dst := "null"
	params.SetDestination(&dst)
	params.SetLimit(&c.limit)
	params.SetOffset(&c.offset)
	sort := "asc"
	params.SetSort(&sort)

	resp, err := c.tonCenterV3Client.API.Blockchain.APIV3GetMessages(
		params,
		c.tonCenterV3Client.Auth,
	)
	if err != nil {
		return nil, err
	}

	return resp.Payload.Messages, nil
}

func (c *TeleportContractLogListener) processMsgs(msgs []*toncenterv3models.Message) {
	for _, msg := range msgs {
		logCell, err := c.extractLogCellFromMsg(msg)
		if err != nil {
			continue
		}
		c.onLogReceived(logCell)
	}
}

func (c *TeleportContractLogListener) extractLogCellFromMsg(msg *toncenterv3models.Message) (*cell.Cell, error) {
	logBytes, err := base64.StdEncoding.DecodeString(msg.MessageContent.Body)
	if err != nil {
		return nil, err
	}

	logCell, err := cell.FromBOC(logBytes)
	if err != nil {
		return nil, err
	}

	return logCell, nil
}
