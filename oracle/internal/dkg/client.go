package dkg

import "fmt"

type Client struct {
	endpoint *Endpoint
}

func CreateClient(endpoint *Endpoint) *Client {
	return &Client{endpoint}
}

func (c *Client) Commit(internalKey []byte) ([]byte, []byte, error) {
	c.endpoint.CommitRequestCh <- &CommitRequest{internalKey}
	result, ok := <-c.endpoint.CommitResultCh
	if !ok {
		return nil, nil, fmt.Errorf("failed to read from CommitResultCh")
	}
	return result.Nonce, result.Commitments, nil
}

func (c *Client) Sign(internalKey []byte, tapTweak []byte, pegoutAddr string, message []byte) ([]byte, error) {
	c.endpoint.SignRequestCh <- &SignRequest{internalKey, tapTweak, pegoutAddr, message}
	result, ok := <-c.endpoint.SignResultCh
	if !ok {
		return nil, fmt.Errorf("failed to read from SignResultCh")
	}
	return result.signingShare, nil
}
