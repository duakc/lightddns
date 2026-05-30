package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"time"

	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"
)

const (
	// ContentTypeUrlEncoded GET
	ContentTypeUrlEncoded = "application/x-www-form-urlencoded"

	// ContentTypeJson ContentTypeFormData POST
	ContentTypeJson     = "application/json; charset=utf-8"
	ContentTypeFormData = "multipart/form-data"
)

const (
	DNSPodServiceName = "dnspod"

	DNSPodActionDescribeRecordList = "DescribeRecordList"

	DNSPodDefaultVersion = "2021-03-23"
)

var tencentCloudEndpoint = mt.Must(urlpkg.Parse("https://dnspod.tencentcloudapi.com"))

type Client struct {
	do httpx.HTTPRequester

	secretId, secretKey string
}

func NewClient(do httpx.HTTPRequester, secretId, secretKey string) *Client {
	return &Client{
		do: do, secretId: secretId, secretKey: secretKey,
	}
}

func (c *Client) DomainInfo(ctx context.Context, search string) (DomainInfo, error) {
	type _domainInfoResponse struct {
		DomainCountInfo struct {
			AllTotal      int `json:"AllTotal"`
			DomainTotal   int `json:"DomainTotal"`
			ErrorTotal    int `json:"ErrorTotal"`
			GroupTotal    int `json:"GroupTotal"`
			LockTotal     int `json:"LockTotal"`
			MineTotal     int `json:"MineTotal"`
			PauseTotal    int `json:"PauseTotal"`
			ShareOutTotal int `json:"ShareOutTotal"`
			ShareTotal    int `json:"ShareTotal"`
			SpamTotal     int `json:"SpamTotal"`
			VipExpire     int `json:"VipExpire"`
			VipTotal      int `json:"VipTotal"`
		} `json:"DomainCountInfo"`

		DomainList []DomainInfo `json:"DomainList"`
	}

	const (
		action   = "DescribeDomainFilterList"
		pageSize = 100
	)

	// fetchPage: one HTTP round-trip. Wrapped in a closure so resp.Body is
	// closed via defer before returning, which keeps the paging loop free of
	// the defer-in-loop trap.
	fetchPage := func(keyword string, offset int) (_domainInfoResponse, error) {
		req := c.newRequest(http.MethodPost)
		req.ExtendHeader.Set("Content-Type", ContentTypeJson)
		body := map[string]any{
			"Type":   "ALL",
			"Limit":  pageSize,
			"Offset": offset,
		}
		if keyword != "" {
			body["Keyword"] = keyword
		}
		req.Body = body

		resp, err := sendRequest(ctx, c.do, req, action, c.secretId, c.secretKey)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			return _domainInfoResponse{}, fmt.Errorf("sendRequest %s: %w", action, err)
		}
		if resp.StatusCode >= 400 {
			return _domainInfoResponse{}, &httpx.BadStatusCodeError{Got: resp.StatusCode}
		}

		var out Response[_domainInfoResponse]
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return _domainInfoResponse{}, fmt.Errorf("decode %s: %w", action, err)
		}
		return out.Data, nil
	}

	// tryKeyword: walk all pages for one keyword; short-circuit on exact match.
	tryKeyword := func(keyword string) (DomainInfo, bool, error) {
		for offset := 0; ; {
			page, err := fetchPage(keyword, offset)
			if err != nil {
				return DomainInfo{}, false, err
			}
			for _, di := range page.DomainList {
				if di.Name == search {
					return di, true, nil
				}
			}
			offset += len(page.DomainList)
			if len(page.DomainList) == 0 || offset >= page.DomainCountInfo.DomainTotal {
				return DomainInfo{}, false, nil
			}
		}
	}

	for _, keyword := range append(domains.CutFromHead(search), "") {
		di, found, err := tryKeyword(keyword)
		if err != nil {
			return DomainInfo{}, err
		}
		if found {
			return di, nil
		}
	}
	return DomainInfo{}, fmt.Errorf("domain not found: %s", search)
}

func (c *Client) newRequest(method string) httpx.ReqConfig {
	req := httpx.NewReqConfig(method, tencentCloudEndpoint)
	return req
}

func sendRequest(ctx context.Context, do httpx.HTTPRequester, req httpx.ReqConfig, action,
	secretId, secretKey string,
) (*http.Response, error) {
	headers := maps.Clone(req.ExtendHeader)
	body, err := httpx.BuildBodyReader(req.Body, headers)
	if err != nil {
		return nil, fmt.Errorf("httpx.BuildBodyReader: %w", err)
	}

	common := Common{
		Action:    action,
		Version:   DNSPodDefaultVersion,
		Timestamp: time.Now().UTC().Unix(),
	}

	commonHeaders := mt.Must(common.Headers())
	httpx.ExtendHeadersOverride(headers, commonHeaders)

	readAllData, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read all data need to signature: %w", err)
	}

	if headers.Get("Content-Length") == "" {
		headers.Set("Content-Length", strconv.Itoa(len(readAllData)))
	}

	sigContext := SigContext{
		Method:  req.Method,
		Headers: headers,

		Body:      readAllData,
		Timestamp: common.Timestamp,
		SecretId:  secretId,
		SecretKey: secretKey,
		Service:   DNSPodServiceName,
	}

	headers.Set("Authorization", mt.Must(sigContext.Authorization()))

	req.ExtendHeader = headers
	req.Body = bytes.NewBuffer(readAllData)

	httpRequest, err := req.ToRequestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("build http request: %w", err)
	}
	return do.Do(httpRequest)
}
