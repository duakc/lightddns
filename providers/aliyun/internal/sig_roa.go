package internal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/duakc/mt/xtypes"

	"github.com/google/uuid"
)

// Aliyun ROA-style V3 signature (ACS3-HMAC-SHA256). Public params travel as
// X-Acs-* request headers, the canonical request hashes the body, and the
// resulting signature lives in the Authorization header.
//
// Kept in the package for reference and for ROA services other than Alidns
// — Alidns itself goes through [AliyunRPCSignClient] (see sig_rpc.go).
//
// https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature

const (
	HeaderAction         = "X-Acs-Action"
	HeaderVersion        = "X-Acs-Version"
	HeaderContentSha256  = "X-Acs-Content-Sha256"
	HeaderDate           = "X-Acs-Date"
	HeaderSignatureNonce = "X-Acs-Signature-Nonce"
	HeaderSecurityToken  = "X-Acs-Security-Token"
	HeaderAuthorization  = "Authorization"
)

const SigAlgoACS3HMACSHA256 = "ACS3-HMAC-SHA256"

// ROACommon collects the per-request public values that V3 transports as
// X-Acs-* headers. Zero fields are filled lazily by [ROACommon.Headers] so
// callers can reuse the struct as a default.
type ROACommon struct {
	SecretSecurityToken string

	Nonce string
	Time  time.Time
}

func (c ROACommon) Headers() http.Header {
	if c.Time.IsZero() {
		c.Time = time.Now().UTC()
	}
	if c.Nonce == "" {
		c.Nonce = uuid.NewString()
	}

	header := make(http.Header)
	header.Set(HeaderDate, c.Time.Format(time.RFC3339))
	header.Set(HeaderSignatureNonce, c.Nonce)
	if c.SecretSecurityToken != "" {
		header.Set(HeaderSecurityToken, c.SecretSecurityToken)
	}
	return header
}

type ROASigContext struct {
	Method string
	Path   string

	Query   urlpkg.Values
	Headers http.Header

	Body []byte

	SecretAccessKeyId     string
	SecretAccessKeySecret string
}

func (sig ROASigContext) CanonicalRequest() (canonicalRequest, signedHeaders, hashedRequestPayload string, err error) {
	if sig.Path == "" {
		sig.Path = "/"
	}
	canonicalURI := xtypes.RFC3986Path(path.Clean(sig.Path)).Encode()

	method := strings.ToUpper(sig.Method)
	queryEnc := xtypes.RFC3986Query{Values: sig.Query}
	canonicalQueryString := queryEnc.Encode()
	hashedRequestPayload = sha256Hex(sig.Body)

	canonicalHeaders, signedHeaders, err := buildHeaders(sig.Headers)
	if err != nil {
		return "", "", "", err
	}

	var b strings.Builder
	b.WriteString(method)
	b.WriteByte('\n')
	b.WriteString(canonicalURI)
	b.WriteByte('\n')
	b.WriteString(canonicalQueryString)
	b.WriteByte('\n')
	b.WriteString(canonicalHeaders)
	b.WriteByte('\n')
	b.WriteString(signedHeaders)
	b.WriteByte('\n')
	b.WriteString(hashedRequestPayload)

	return b.String(), signedHeaders, hashedRequestPayload, nil
}

func (sig ROASigContext) StringToSign(canonicalRequest string) string {
	return strings.Join([]string{
		SigAlgoACS3HMACSHA256, sha256Hex([]byte(canonicalRequest)),
	}, "\n")
}

func (sig ROASigContext) BuildAuthorization(stringToSign, signedHeaders string) string {
	signature := hex.EncodeToString(hmacsha256([]byte(sig.SecretAccessKeySecret), stringToSign))

	var b strings.Builder
	b.WriteString(SigAlgoACS3HMACSHA256)
	b.WriteByte(' ')
	b.WriteString("Credential=")
	b.WriteString(sig.SecretAccessKeyId)
	b.WriteByte(',')
	b.WriteString("SignedHeaders=")
	b.WriteString(signedHeaders)
	b.WriteByte(',')
	b.WriteString("Signature=")
	b.WriteString(strings.ToLower(signature))
	return b.String()
}

func (sig ROASigContext) Authorization() (string, error) {
	canonicalRequest, signedHeaders, _, err := sig.CanonicalRequest()
	if err != nil {
		return "", err
	}
	stringToSign := sig.StringToSign(canonicalRequest)
	return sig.BuildAuthorization(stringToSign, signedHeaders), nil
}

func buildHeaders(headers http.Header) (canonicalHeaders, signedHeaders string, err error) {
	signed := make(map[string]string)
	headerSorted := make([]string, 0, 8)

	for k, vals := range headers {
		lower := strings.ToLower(strings.TrimSpace(k))
		if !(lower == "host" ||
			strings.HasPrefix(lower, "x-acs-") ||
			lower == "content-type") {

			continue
		}

		val := strings.TrimSpace(strings.Join(vals, ","))
		headerSorted = append(headerSorted, lower)
		signed[lower] = val
	}

	if _, ok := signed["host"]; !ok {
		return "", "", fmt.Errorf("missing header 'Host'")
	}

	sort.Strings(headerSorted)

	var b strings.Builder
	for _, k := range headerSorted {
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(signed[k])
		b.WriteByte('\n')
	}
	return b.String(), strings.Join(headerSorted, ";"), nil
}

func hmacsha256(key []byte, toSignString string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(toSignString))
	return h.Sum(nil)
}

func sha256Hex(byteArray []byte) string {
	if len(byteArray) == 0 {
		// a fast path to reduce calc
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}
	hash := sha256.New()
	_, _ = hash.Write(byteArray)
	hexString := hex.EncodeToString(hash.Sum(nil))

	return hexString
}
