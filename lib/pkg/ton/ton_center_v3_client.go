package ton

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3client"
)

type TonCenterV3Client struct {
	API  *toncenterv3client.TonCenterV3
	Auth runtime.ClientAuthInfoWriter
}

func NewTonCenterV3Client(host string, apiKey string, basePath string, scheme string, debug bool) (
	*TonCenterV3Client,
	error,
) {
	if host == "" {
		host = toncenterv3client.DefaultHost
	}

	if basePath == "" {
		basePath = toncenterv3client.DefaultBasePath
	}

	schemes := toncenterv3client.DefaultSchemes
	if scheme != "" {
		schemes = []string{scheme}
	}

	transport := httptransport.New(host, basePath, schemes)

	if debug == true {
		transport.SetDebug(true)
	}

	auth := httptransport.APIKeyAuth(
		"X-Api-Key",
		"header",
		apiKey,
	)

	api := toncenterv3client.New(transport, strfmt.Default)

	return &TonCenterV3Client{
		API:  api,
		Auth: auth,
	}, nil
}
