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

	"github.com/duakc/mt/xtypes"
)

// https://help.aliyun.com/zh/sdk/product-overview/v3-request-structure-and-signature

type SigContext struct {
	Method string
	Path   string

	Query   urlpkg.Values
	Headers http.Header

	Body []byte

	SecretAccessKeyId     string
	SecretAccessKeySecret string
}

func (sig SigContext) CanonicalRequest() (
	canonicalRequest,
	hashedRequestPayload,
	signedHeaders string,

	err error,
) {
	if sig.Path == "" {
		sig.Path = "/"
	}
	canonicalURI := xtypes.RFC3986Path(path.Clean(sig.Path)).Encode()

	var canonicalQueryString string

	method := strings.ToUpper(sig.Method)
	switch method {
	case http.MethodGet:
		hashedRequestPayload = sha256Hex(nil)
		encoder := xtypes.RFC3986Query{Values: sig.Query}
		canonicalQueryString = encoder.Encode()
	case http.MethodPost:
		hashedRequestPayload = sha256Hex(sig.Body)
		canonicalQueryString = sha256Hex(nil)
	}

	var canonicalHeaders string
	canonicalHeaders, signedHeaders, err = buildHeaders(sig.Headers)
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

	return b.String(), hashedRequestPayload, signedHeaders, nil
}

func (sig SigContext) StringToSign(canonicalRequest string) string {
	return strings.Join([]string{
		SigAlgoACS3HMACSHA256, sha256Hex([]byte(canonicalRequest)),
	}, "\n")
}

func (sig SigContext) BuildAuthorization(stringToSign, signedHeaders string) string {
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
	b.WriteString(signature)
	return b.String()
}

func (sig SigContext) Header() (http.Header, error) {
	canonicalRequest, hashedRequestPayload, signedHeaders, err := sig.CanonicalRequest()
	if err != nil {
		return nil, err
	}
	stringToSign := sig.StringToSign(canonicalRequest)
	authorization := sig.BuildAuthorization(stringToSign, signedHeaders)
	header := make(http.Header)
	header.Set(HeaderAuthorization, authorization)
	header.Set(HeaderContentSha256, hashedRequestPayload)

	return header, nil
}

func buildHeaders(headers http.Header) (canonicalHeaders, signedHeaders string, err error) {
	signed := make(map[string]string)
	headerSorted := make([]string, 0, 3)

	for k, vals := range headers {
		lower := strings.ToLower(strings.TrimSpace(k))
		if lower == "" ||
			(lower != "content-type" && lower != "host" &&
				!strings.HasPrefix(lower, "x-acs-")) {
			continue
		}

		val := strings.TrimSpace(strings.Join(vals, ","))
		headerSorted = append(headerSorted, lower)
		signed[lower] = val
	}

	if _, ok := signed["host"]; !ok {
		return "", "", fmt.Errorf("missing header 'Host'")
	}

	if _, ok := signed["content-type"]; !ok {
		return "", "", fmt.Errorf("missing header 'Content-Type'")
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
