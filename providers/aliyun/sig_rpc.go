package aliyun

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	urlpkg "net/url"
	"strings"

	"github.com/duakc/mt/xtypes"
)

// Aliyun RPC v1 signature (SignatureVersion=1.0, SignatureMethod=HMAC-SHA1).
//
// All public params (Action, Version, AccessKeyId, Timestamp,
// SignatureNonce, SignatureMethod, SignatureVersion, Format) and per-action
// params travel on the URL query string; the computed Signature is also
// appended to the query string.
//
// Example URL shape (from Aliyun docs):
//
//	http://ecs.aliyuncs.com/?SignatureVersion=1.0&Action=DescribeDedicatedHosts
//	  &Format=JSON&SignatureNonce=…&Version=2014-05-26&AccessKeyId=…
//	  &Signature=…&SignatureMethod=HMAC-SHA1&Timestamp=…&RegionId=…
//
// https://help.aliyun.com/document_detail/315526.html

// RPC v1 public-parameter names. These are the URL query keys the
// signature spec reserves; per-action keys (DomainName, RR, …) must not
// collide with them.
const (
	ParamAccessKeyId      = "AccessKeyId"
	ParamAction           = "Action"
	ParamFormat           = "Format"
	ParamSecurityToken    = "SecurityToken"
	ParamSignature        = "Signature"
	ParamSignatureMethod  = "SignatureMethod"
	ParamSignatureNonce   = "SignatureNonce"
	ParamSignatureVersion = "SignatureVersion"
	ParamTimestamp        = "Timestamp"
	ParamVersion          = "Version"
)

// Fixed values for the RPC v1 signing flow.
const (
	RPCSignatureMethod  = "HMAC-SHA1"
	RPCSignatureVersion = "1.0"
	RPCFormatJSON       = "JSON"
)

type RPCSigContext struct {
	// Method is the HTTP verb (GET / POST). Uppercased before signing.
	Method string

	// Query is the full URL query — public params already merged in, but
	// without "Signature". The signer reads it but does not mutate it.
	Query urlpkg.Values

	// SecretAccessKeySecret is the Aliyun AccessKey secret. The HMAC key is
	// `SecretAccessKeySecret + "&"` (the trailing '&' is mandated by the
	// spec for v1).
	SecretAccessKeySecret string
}

func (sig RPCSigContext) CanonicalQueryString() string {
	encoder := xtypes.RFC3986Query{Values: sig.Query}
	return encoder.Encode()
}

func (sig RPCSigContext) StringToSign(canonicalQueryString string) string {
	var b strings.Builder
	b.WriteString(strings.ToUpper(sig.Method))
	b.WriteByte('&')
	b.WriteString(rpcPercentEncode("/"))
	b.WriteByte('&')
	b.WriteString(rpcPercentEncode(canonicalQueryString))
	return b.String()
}

func (sig RPCSigContext) Signature() string {
	canonical := sig.CanonicalQueryString()
	stringToSign := sig.StringToSign(canonical)

	mac := hmac.New(sha1.New, []byte(sig.SecretAccessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func rpcPercentEncode(s string) string {
	encoded := urlpkg.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
