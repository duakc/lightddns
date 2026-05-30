package internal

import (
	"context"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/httpxx"

	"github.com/duakc/mt"
)

const (
	// ContentTypeUrlEncoded GET
	ContentTypeUrlEncoded = "application/x-www-form-urlencoded"

	// ContentTypeJson ContentTypeFormData POST
	ContentTypeJson     = "application/json"
	ContentTypeFormData = "multipart/form-data"
)

const (
	DNSPodServiceName = "dnspod"

	DNSPodActionDescribeRecordList = "DescribeRecordList"
)

var tencentCloudEndpoint = mt.Must(urlpkg.Parse("https://dnspod.tencentcloudapi.com"))

type Client struct {
	do httpxx.HTTPRequester

	secretId, secretKey string
}

func NewClient(do httpxx.HTTPRequester, secretId, secretKey string) *Client {
	return &Client{}
}

func (c *Client) DescribeRecordList(ctx context.Context, domain string) {
}

func (c *Client) DescribeRecordListFilter(ctx context.Context, domain string,
	recordType []string) {
}

func (c *Client) newRequest(method, action string) httpxx.ReqConfig {
	req := httpxx.ReqConfig{
		Method: method,
	}
	req.ExtendHeader = http.Header{
		"X-TC-Action":  []string{action},
		"X-TC-Version": []string{"2021-03-23"},
	}
	return req
}
