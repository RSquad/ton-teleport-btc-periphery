package loglistener

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3client/blockchain"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3models"
)

type LogListener struct {
	tonCenterV3Client *ton.TonCenterV3Client
	listenAddr        *address.Address
	offset            int64
	limit             int64
	onLogReceived     func(*cell.Cell, string, time.Time)
}

func New(
	tonCenterV3Client *ton.TonCenterV3Client,
	listenAddr *address.Address,
	onLogReceived func(*cell.Cell, string, time.Time),
) (
	*LogListener,
	error,
) {
	return &LogListener{
		tonCenterV3Client,
		listenAddr,
		0,
		128,
		onLogReceived,
	}, nil
}

func (c *LogListener) StartListen() {
	log.Printf("[LogListener] listening for %v started", c.listenAddr.String())
	c.listen()
}

func (c *LogListener) listen() {
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

func (c *LogListener) fetchMsgs() ([]*toncenterv3models.Message, error) {
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

func (c *LogListener) processMsgs(msgs []*toncenterv3models.Message) {
	for _, msg := range msgs {
		logCell, err := c.extractLogCellFromMsg(msg)
		if err != nil {
			continue
		}
		unixTime, err := strconv.ParseInt(msg.CreatedAt, 10, 64)
		if err != nil {
			log.Printf("[LogListener] failed to parse msg creation time: %v", err)
			continue
		}
		createdAt := time.Unix(unixTime, 0)
		c.onLogReceived(logCell, msg.MessageContent.Hash, createdAt)
	}
}

func (c *LogListener) extractLogCellFromMsg(msg *toncenterv3models.Message) (*cell.Cell, error) {
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
