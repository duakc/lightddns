package internal

import (
	urlpkg "net/url"

	"github.com/duakc/mt"
)

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
