package internal

// https://cloud.tencent.com/document/api/1427/56189

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"sort"
	"strings"
	"time"

	"github.com/duakc/mt/xtypes"
)

const sigAlgo = "TC3-HMAC-SHA256"

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func hmacsha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

const sigCanonicalURI = '/'

type SigContext struct {
	// Method is GET or POST
	Method string

	// All Headers will be encoded
	Headers http.Header

	// It will be encoded by xtypes.RFC3986Query,
	// as tencent cloud platform need a strict RFC3986 style url encoding.
	Query urlpkg.Values

	Body []byte

	// StringToSign
	// SigMethod string // fixed TC3-HMAC-SHA256 now

	// This value should equal with X-TC-Timestamp Header
	Timestamp int64

	//
	SecretId  string
	SecretKey string

	// In this project,  this value is fixed, DNSPodServiceName.
	Service string
}

func (ctx SigContext) CanonicalRequest() (string, string, error) {
	var (
		method      = strings.ToUpper(ctx.Method)
		queryStr    string
		payloadHash string
	)
	switch method {
	case http.MethodGet:
		encoder := xtypes.RFC3986Query{Values: ctx.Query}
		queryStr = encoder.Encode()
		payloadHash = sha256hex("")
	case http.MethodPost:
		body := ctx.Body
		if body == nil {
			body = []byte{}
		}
		payloadHash = sha256hex(string(body))
	default:
		return "", "", fmt.Errorf("unsupported method: %s", method)
	}

	canonicalHeaders, signedHeaders, err := buildHeaders(ctx.Headers)
	if err != nil {
		return "", "", err
	}

	var b strings.Builder
	b.WriteString(method)
	b.WriteByte('\n')
	b.WriteByte(sigCanonicalURI)
	b.WriteByte('\n')
	b.WriteString(queryStr)
	b.WriteByte('\n')
	b.WriteString(canonicalHeaders)
	b.WriteByte('\n')
	b.WriteString(signedHeaders)
	b.WriteByte('\n')
	b.WriteString(payloadHash)

	return b.String(), signedHeaders, nil
}

func (ctx SigContext) StringToSign(canonicalReq string) (stringToSign string, credentialScope string) {
	ts := ctx.Timestamp
	date := time.Unix(ts, 0).UTC().Format("2006-01-02")
	credentialScope = date + "/" + ctx.Service + "/tc3_request"
	hashedReq := sha256hex(canonicalReq)

	var b strings.Builder
	b.WriteString(sigAlgo)
	b.WriteByte('\n')
	_, _ = fmt.Fprint(&b, ts)
	b.WriteByte('\n')
	b.WriteString(credentialScope)
	b.WriteByte('\n')
	b.WriteString(hashedReq)

	return b.String(), credentialScope
}

// BuildAuthorization
//   - stringToSign  StringToSign()
//   - credentialScope  StringToSign()
//   - signedHeaders  CanonicalRequest()
func (ctx SigContext) BuildAuthorization(stringToSign, credentialScope, signedHeaders string) string {
	date := time.Unix(ctx.Timestamp, 0).UTC().Format("2006-01-02")

	secretDate := hmacsha256(date, "TC3"+ctx.SecretKey)
	secretService := hmacsha256(ctx.Service, secretDate)
	secretSigning := hmacsha256("tc3_request", secretService)
	signatureBytes := hmacsha256(stringToSign, secretSigning)
	signature := hex.EncodeToString([]byte(signatureBytes))

	var b strings.Builder
	b.WriteString(sigAlgo)
	b.WriteString(" Credential=")
	b.WriteString(ctx.SecretId)
	b.WriteByte('/')
	b.WriteString(credentialScope)
	b.WriteString(", SignedHeaders=")
	b.WriteString(signedHeaders)
	b.WriteString(", Signature=")
	b.WriteString(signature)

	return b.String()
}

func (ctx SigContext) Authorization() (string, error) {
	canonicalReq, signedHeaders, err := ctx.CanonicalRequest()
	if err != nil {
		return "", err
	}

	stringToSign, credentialScope := ctx.StringToSign(canonicalReq)

	v := ctx.BuildAuthorization(stringToSign, credentialScope, signedHeaders)
	return v, nil
}

func buildHeaders(headers http.Header) (canonicalHeaders, signedHeaders string, err error) {
	allowed := map[string]bool{"content-type": true, "host": true, "x-tc-action": true}

	signed := make(map[string]string)
	headerSorted := make([]string, 0, 3)

	for k, vals := range headers {
		lower := strings.ToLower(strings.TrimSpace(k))
		if !allowed[lower] {
			continue
		}
		val := strings.TrimSpace(strings.Join(vals, ","))
		if lower == "x-tc-action" {
			val = strings.ToLower(val)
		}
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
