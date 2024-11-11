package loglistener

import (
	"context"
	"log"
	"time"

	"github.com/xssnick/tonutils-go/address"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3client/blockchain"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3models"
)

type LogListenerInterface interface {
    StartListen(ctx context.Context) error
    StopListen()
}

type LogListener struct {
    tonCenterV3Client *ton.TonCenterV3Client
    addrToListen      *address.Address
    offset            int64
    limit             int64
    cancelFunc        context.CancelFunc
}

func NewLogListener(tonCenterV3Client *ton.TonCenterV3Client, addrToListen *address.Address) (
    LogListenerInterface,
    error,
) {
    logListener := &LogListener{
        tonCenterV3Client: tonCenterV3Client,
        addrToListen:      addrToListen,
        offset:            0,
        limit:             128,
    }

    return logListener, nil
}

func (c *LogListener) StartListen(ctx context.Context) error {
    log.Println("[LogListener] listening started")

    ctx, cancel := context.WithCancel(ctx)
    c.cancelFunc = cancel

    err := c.listen(ctx)
    if err != nil {
        return err
    }

    return nil
}

func (c *LogListener) StopListen() {
    if c.cancelFunc != nil {
        c.cancelFunc()
    }
}

func (c *LogListener) listen(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            log.Println("[LogListener] listening stopped")
            return nil
        default:
            logs, err := c.fetchLogs()
            if err != nil {
                return err
            }

            log.Println(len(logs.Messages))

            c.offset += int64(len(logs.Messages))

            if int64(len(logs.Messages)) == c.limit {
                return c.listen(ctx)
            }

            log.Println("[LogListener] max logs fetched, waiting 10 seconds before retrying...")
            time.Sleep(10 * time.Second)
        }
    }
}

func (c *LogListener) fetchLogs() (*toncenterv3models.MessagesResponse, error) {
    params := blockchain.NewAPIV3GetMessagesParamsWithTimeout(30 * time.Second)
    src := c.addrToListen.String()
    params.SetSource(&src)
    params.SetLimit(&c.limit)
    params.SetOffset(&c.offset)

    resp, err := c.tonCenterV3Client.API.Blockchain.APIV3GetMessages(
        params,
        c.tonCenterV3Client.Auth,
    )

    if err != nil {
        return nil, err
    }

    return resp.Payload, nil
}
