package internal

import (
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"
)

// https://help.aliyun.com/zh/dns/api-alidns-2015-01-09-overview

const (
	HeaderAction  = "X-Acs-Action"
	HeaderVersion = "X-Acs-Version"

	HeaderContentSha256  = "X-Acs-Content-Sha256"
	HeaderDate           = "X-Acs-Date"
	HeaderSignatureNonce = "X-Acs-Signature-Nonce"
	HeaderAuthorization  = "Authorization"
	HeaderSecurityToken  = "X-Acs-Security-Token"
)

const (
	SigAlgoACS3HMACSHA256 = "ACS3-HMAC-SHA256"
)

const (
	ContentTypeJSON         = "application/json"
	ContentTypeBinaryStream = "application/octet-stream"
	ContentTypeForm         = "application/x-www-form-urlencoded"
)

var AliyunDNSEndpoint = mt.Must(urlpkg.Parse("https://alidns.aliyuncs.com"))

type Client struct {
	do httpx.HTTPRequester
}

func NewClient(do httpx.HTTPRequester,
	secretAccessKeyId, secretAccessKeySecret, secretSecurityToken string,
) *Client {
	return &Client{
		do: &AliyunSignClient{
			HTTPRequester:         do,
			SecretSecurityToken:   secretSecurityToken,
			SecretAccessKeyId:     secretAccessKeyId,
			SecretAccessKeySecret: secretAccessKeySecret,
		},
	}
}

func newRequest(action, version string) httpx.ReqConfig {
	// aliyun dns api use RPC style, we use POST method as default for
	// a better compatibility and sync with tencentcloud api.

	r := httpx.NewReqConfig(http.MethodPost, AliyunDNSEndpoint)
	r.ExtendHeader.Set(HeaderAction, action)
	r.ExtendHeader.Set(HeaderVersion, version)
	return r
}

func newGetRequest(action, version string) httpx.ReqConfig {
	r := httpx.NewReqConfig(http.MethodGet, AliyunDNSEndpoint)
	r.ExtendHeader.Set(HeaderAction, action)
	r.ExtendHeader.Set(HeaderVersion, version)
	return r
}
