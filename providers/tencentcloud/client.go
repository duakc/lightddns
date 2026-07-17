package tencentcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	urlpkg "net/url"
	"strings"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

// Tencent Cloud DNSPod API surface — we only use the v3 API.
//
// https://cloud.tencent.com/document/api/1427/56188
const (
	DNSPodServiceName    = "dnspod"
	DNSPodDefaultVersion = "2021-03-23"

	DNSPodActionDescribeDomainFilterList = "DescribeDomainFilterList"
	DNSPodActionDescribeRecordList       = "DescribeRecordList"
	DNSPodActionCreateRecord             = "CreateRecord"
	DNSPodActionModifyRecord             = "ModifyRecord"
	DNSPodActionDeleteRecord             = "DeleteRecord"

	// DNSPodErrCodeNoDataOfRecord is returned by DescribeRecordList when the
	// domain exists but has no records matching the filter. For DDNS this is
	// the normal "first run" state, not an error — callers should treat it
	// as an empty list.
	DNSPodErrCodeNoDataOfRecord = "ResourceNotFound.NoDataOfRecord"
)

const (
	HeaderAction    = "X-TC-Action"
	HeaderVersion   = "X-TC-Version"
	HeaderTimestamp = "X-TC-Timestamp"
	HeaderToken     = "X-TC-Token"
	HeaderLanguage  = "X-TC-Language"
	HeaderRegion    = "X-TC-Region"
)

const ContentTypeJSON = "application/json; charset=utf-8"

var TencentCloudEndpoint = mt.Must(urlpkg.Parse("https://dnspod.tencentcloudapi.com"))

var (
	_ ddnsx.DDNSClient[Record] = (*Client)(nil)
	_ ddnsx.ZoneSearcher       = (*Client)(nil)
)

type Client struct {
	logger *zap.Logger
	do     httpx.HTTPRequester

	zones ddnsx.ZoneCache
}

func NewClient(logger *zap.Logger, do httpx.HTTPRequester, secretId, secretKey string) *Client {
	return &Client{
		logger: logger,
		do: &TencentSignHTTPRequester{
			HTTPRequester: do,
			Logger:        logger,
			SecretId:      secretId,
			SecretKey:     secretKey,
			Service:       DNSPodServiceName,
		},
	}
}

// DescribeDomainFilterList — https://cloud.tencent.com/document/api/1427/56173
func (c *Client) DescribeDomainFilterList(ctx context.Context,
	req DescribeDomainFilterListRequest,
) (resp DescribeDomainFilterListResponse, err error) {
	return doAction[DescribeDomainFilterListResponse](ctx, c, DNSPodActionDescribeDomainFilterList, req)
}

// DescribeRecordList — https://cloud.tencent.com/document/api/1427/56166
func (c *Client) DescribeRecordList(ctx context.Context,
	req DescribeRecordListRequest,
) (resp DescribeRecordListResponse, err error) {
	return doAction[DescribeRecordListResponse](ctx, c, DNSPodActionDescribeRecordList, req)
}

// CreateRecord — https://cloud.tencent.com/document/api/1427/56180
func (c *Client) CreateRecord(ctx context.Context,
	req CreateRecordRequest,
) (resp CreateRecordResponse, err error) {
	return doAction[CreateRecordResponse](ctx, c, DNSPodActionCreateRecord, req)
}

// ModifyRecord — https://cloud.tencent.com/document/api/1427/56157
func (c *Client) ModifyRecord(ctx context.Context,
	req ModifyRecordRequest,
) (resp ModifyRecordResponse, err error) {
	return doAction[ModifyRecordResponse](ctx, c, DNSPodActionModifyRecord, req)
}

// DeleteRecord — https://cloud.tencent.com/document/api/1427/56176
func (c *Client) DeleteRecord(ctx context.Context,
	req DeleteRecordRequest,
) (resp DeleteRecordResponse, err error) {
	return doAction[DeleteRecordResponse](ctx, c, DNSPodActionDeleteRecord, req)
}

func (c *Client) ResolveZone(ctx context.Context, fqdn string) (ddnsx.Zone, error) {
	return c.zones.Resolve(ctx, fqdn, c)
}

func (c *Client) SearchZones(ctx context.Context, keyword ddnsx.ZoneName) ([]ddnsx.Zone, error) {
	const pageSize = 100

	logger := c.actionLogger(DNSPodActionDescribeDomainFilterList).
		With(zap.Stringer("keyword", keyword))
	logger.Info("search domain id from upstream")

	var zones []ddnsx.Zone
	for offset := 0; ; {
		page, err := c.DescribeDomainFilterList(ctx, DescribeDomainFilterListRequest{
			Type:    "ALL",
			Limit:   pageSize,
			Offset:  offset,
			Keyword: keyword.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeDomainFilterList: %w", err)
		}
		for _, domain := range page.DomainList {
			if !domains.IsDomainName(domain.Name) {
				continue
			}
			zones = append(zones, ddnsx.Zone{Name: ddnsx.ZoneName(domain.Name)})
		}
		offset += len(page.DomainList)
		if len(page.DomainList) == 0 || offset >= page.DomainCountInfo.DomainTotal {
			return zones, nil
		}
	}
}

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ddnsx.Existing[Record], error) {
	const pageSize = 100

	subdomain, err := relativeSubDomain(key.FQDN, key.Zone.Name.String())
	if err != nil {
		return nil, err
	}

	var records []ddnsx.Existing[Record]
	for offset := 0; ; {
		page, err := c.DescribeRecordList(ctx, DescribeRecordListRequest{
			Domain:     key.Zone.Name.String(),
			Subdomain:  subdomain,
			RecordType: key.Type.String(),
			Limit:      pageSize,
			Offset:     offset,
		})
		if err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.Code == DNSPodErrCodeNoDataOfRecord {
				return records, nil
			}
			return nil, fmt.Errorf("DescribeRecordList: %w", err)
		}
		for _, record := range page.RecordList {
			if record.Type != key.Type.String() {
				continue
			}
			address, err := netip.ParseAddr(record.Value)
			if err != nil {
				return nil, fmt.Errorf("record %d: not an address: %s: %w", record.RecordId, record.Value, err)
			}
			records = append(records, ddnsx.Existing[Record]{
				Addr:   address.Unmap(),
				Record: record,
			})
		}
		offset += len(page.RecordList)
		if len(page.RecordList) == 0 || offset >= page.RecordCountInfo.TotalCount {
			return records, nil
		}
	}
}

func (c *Client) Create(ctx context.Context, target ddnsx.RecordTarget) error {
	subdomain, err := relativeSubDomain(target.FQDN, target.Zone.Name.String())
	if err != nil {
		return err
	}
	_, err = c.CreateRecord(ctx, CreateRecordRequest{
		Domain:     target.Zone.Name.String(),
		SubDomain:  subdomain,
		RecordType: target.Type.String(),
		RecordLine: DefaultRecordLine,
		Value:      target.Address.Unmap().String(),
		TTL:        target.TTL,
	})
	return err
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordTarget, record Record) error {
	subdomain, err := relativeSubDomain(target.FQDN, target.Zone.Name.String())
	if err != nil {
		return err
	}
	_, err = c.ModifyRecord(ctx, ModifyRecordRequest{
		Domain:     target.Zone.Name.String(),
		RecordId:   record.RecordId,
		SubDomain:  subdomain,
		RecordType: target.Type.String(),
		RecordLine: lineOrDefault(record.Line),
		Value:      target.Address.Unmap().String(),
		TTL:        target.TTL,
	})
	return err
}

func (c *Client) Delete(ctx context.Context, key ddnsx.RecordKey, record Record) error {
	_, err := c.DeleteRecord(ctx, DeleteRecordRequest{
		Domain:   key.Zone.Name.String(),
		RecordId: record.RecordId,
	})
	return err
}

func lineOrDefault(line string) string {
	if line == "" {
		return DefaultRecordLine
	}
	return line
}

func relativeSubDomain(fqdn, zone string) (string, error) {
	fqdn, err := ddnsx.NormalizeFQDN(fqdn)
	if err != nil {
		return "", err
	}
	zone, err = ddnsx.NormalizeFQDN(zone)
	if err != nil {
		return "", err
	}
	if fqdn == zone {
		return "@", nil
	}
	if subdomain, found := strings.CutSuffix(fqdn, "."+zone); found {
		return subdomain, nil
	}
	return "", fmt.Errorf("fqdn %q is not within zone %q", fqdn, zone)
}

func (c *Client) actionLogger(action string) *zap.Logger {
	return c.logger.With(zap.String("action", action))
}

func doAction[Resp any](ctx context.Context, c *Client, action string, body any) (Resp, error) {
	var zero Resp

	logger := c.actionLogger(action)

	req := newRequest(http.MethodPost, action)
	req.ExtendHeader.Set("Host", TencentCloudEndpoint.Host)
	req.ExtendHeader.Set("Content-Type", ContentTypeJSON)
	req.Body = body

	httpReq, err := req.ToRequestContext(ctx)
	if err != nil {
		return zero, fmt.Errorf("build request %s: %w", action, err)
	}

	resp, err := c.do.Do(httpReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return zero, fmt.Errorf("send %s: %w", action, err)
	}
	if resp == nil {
		panic("empty response with no error")
	}
	if resp.StatusCode >= 400 {
		return zero, &httpx.BadStatusCodeError{Got: resp.StatusCode}
	}

	var out Response[Resp]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logger.Warn("decode response failed", zap.Error(err))
		return zero, fmt.Errorf("decode %s: %w", action, err)
	}

	if out.Error != nil {
		logger.Warn("api returned error",
			zap.String("error_code", out.Error.Code),
			zap.String("error_message", out.Error.Message),
			zap.String("request_id", out.RequestID))
		return zero, out.Error
	}
	return out.Data, nil
}

func newRequest(method, action string) httpx.ReqConfig {
	req := httpx.NewReqConfig(method, TencentCloudEndpoint)
	req.ExtendHeader.Set(HeaderAction, action)
	req.ExtendHeader.Set(HeaderVersion, DNSPodDefaultVersion)
	return req
}
