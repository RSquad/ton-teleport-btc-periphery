package loglistener

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"k8s.io/client-go/util/workqueue"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3client/blockchain"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3models"
)

type LogListenerInterface interface {
	StartListen()
}

type LogListener struct {
	tonCenterV3Client *ton.TonCenterV3Client
	listenAddr        *address.Address
	offset            int64
	limit             int64
	outQueue          *workqueue.Typed[*cell.Cell]
}

func NewLogListener(
	tonCenterV3Client *ton.TonCenterV3Client,
	listenAddr *address.Address,
	outQueue *workqueue.Typed[*cell.Cell],
) (
	LogListenerInterface,
	error,
) {
	return &LogListener{
		tonCenterV3Client: tonCenterV3Client,
		listenAddr:        listenAddr,
		offset:            0,
		limit:             1000,
		outQueue:          outQueue,
	}, nil
}

func (c *LogListener) StartListen() {
	log.Println("[LogListener] listening started")
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
		c.outQueue.Add(logCell)
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
