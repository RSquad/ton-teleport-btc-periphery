package ton

import (
	"os"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/generated/toncenterv3client"
)

type TonCenterV3Client struct {
	API  *toncenterv3client.TonCenterV3
	Auth runtime.ClientAuthInfoWriter
}

func NewTonCenterV3Client(debug bool) (
	*TonCenterV3Client,
	error,
) {
	host := os.Getenv("COMMON_TON_CENTER_V3_HOST")
	if host == "" {
		host = toncenterv3client.DefaultHost
	}

	basePath := os.Getenv("COMMON_TON_CENTER_V3_BASE_PATH")
	if basePath == "" {
		basePath = toncenterv3client.DefaultBasePath
	}

	scheme := os.Getenv("COMMON_TON_CENTER_V3_SCHEME")
	schemes := toncenterv3client.DefaultSchemes
	if scheme != "" {
		schemes = []string{scheme}
	}

	apiKey := os.Getenv("COMMON_TON_CENTER_API_KEY")

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
