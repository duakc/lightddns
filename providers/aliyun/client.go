package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
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

// Aliyun Alidns API surface — RPC-style, signed with v1 (HMAC-SHA1).
// See sig_rpc.go for the signing logic and sig_roa.go for the ROA / V3
// scheme kept around for non-Alidns services.
//
// https://help.aliyun.com/zh/dns/api-alidns-2015-01-09-overview
const (
	AlidnsServiceName    = "alidns"
	AlidnsDefaultVersion = "2015-01-09"

	AlidnsActionDescribeDomains       = "DescribeDomains"
	AlidnsActionDescribeDomainRecords = "DescribeDomainRecords"
	AlidnsActionAddDomainRecord       = "AddDomainRecord"
	AlidnsActionUpdateDomainRecord    = "UpdateDomainRecord"
	AlidnsActionDeleteDomainRecord    = "DeleteDomainRecord"
)

var AliyunDNSEndpoint = mt.Must(urlpkg.Parse("https://alidns.aliyuncs.com"))

var (
	_ ddnsx.DDNSClient[Record] = (*Client)(nil)
	_ ddnsx.ZoneSearcher       = (*Client)(nil)
)

type Client struct {
	logger *zap.Logger
	do     httpx.HTTPRequester

	zones ddnsx.ZoneCache
}

func NewClient(logger *zap.Logger, do httpx.HTTPRequester,
	secretAccessKeyId, secretAccessKeySecret, secretSecurityToken string,
) *Client {
	return &Client{
		logger: logger,
		// Alidns uses the RPC-style v1 signature: all public params (and
		// the resulting Signature) ride on the URL query string. The
		// ROA-style signer in [AliyunROASignClient] is kept in the package
		// for reference / future use against ROA services.
		do: &AliyunRPCSignClient{
			HTTPRequester:         do,
			Logger:                logger,
			SecretSecurityToken:   secretSecurityToken,
			SecretAccessKeyId:     secretAccessKeyId,
			SecretAccessKeySecret: secretAccessKeySecret,
		},
	}
}

// DescribeDomains — https://help.aliyun.com/document_detail/29751.html
func (c *Client) DescribeDomains(ctx context.Context,
	req DescribeDomainsRequest,
) (resp DescribeDomainsResponse, err error) {
	return doAction[DescribeDomainsResponse](ctx, c, AlidnsActionDescribeDomains, req.Query())
}

// DescribeDomainRecords — https://help.aliyun.com/document_detail/29774.html
func (c *Client) DescribeDomainRecords(ctx context.Context,
	req DescribeDomainRecordsRequest,
) (resp DescribeDomainRecordsResponse, err error) {
	return doAction[DescribeDomainRecordsResponse](ctx, c, AlidnsActionDescribeDomainRecords, req.Query())
}

// AddDomainRecord — https://help.aliyun.com/document_detail/29772.html
func (c *Client) AddDomainRecord(ctx context.Context,
	req AddDomainRecordRequest,
) (resp AddDomainRecordResponse, err error) {
	return doAction[AddDomainRecordResponse](ctx, c, AlidnsActionAddDomainRecord, req.Query())
}

// UpdateDomainRecord — https://help.aliyun.com/document_detail/29773.html
func (c *Client) UpdateDomainRecord(ctx context.Context,
	req UpdateDomainRecordRequest,
) (resp UpdateDomainRecordResponse, err error) {
	return doAction[UpdateDomainRecordResponse](ctx, c, AlidnsActionUpdateDomainRecord, req.Query())
}

// DeleteDomainRecord — https://help.aliyun.com/document_detail/29771.html
func (c *Client) DeleteDomainRecord(ctx context.Context,
	req DeleteDomainRecordRequest,
) (resp DeleteDomainRecordResponse, err error) {
	return doAction[DeleteDomainRecordResponse](ctx, c, AlidnsActionDeleteDomainRecord, req.Query())
}

func (c *Client) ResolveZone(ctx context.Context, fqdn string) (ddnsx.Zone, error) {
	return c.zones.Resolve(ctx, fqdn, c)
}

func (c *Client) SearchZones(ctx context.Context, keyword ddnsx.ZoneName) ([]ddnsx.Zone, error) {
	const pageSize = 100

	logger := c.actionLogger(AlidnsActionDescribeDomains).
		With(zap.Stringer("keyword", keyword))
	logger.Info("search domain from upstream")

	var zones []ddnsx.Zone
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.DescribeDomains(ctx, DescribeDomainsRequest{
			KeyWord:    keyword.String(),
			PageNumber: pageNumber,
			PageSize:   pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeDomains: %w", err)
		}
		for _, domain := range page.Domains.Domain {
			if !domains.IsDomainName(domain.DomainName) {
				continue
			}
			zones = append(zones, ddnsx.Zone{Name: ddnsx.ZoneName(domain.DomainName)})
		}
		if len(page.Domains.Domain) < pageSize ||
			int64(pageNumber)*pageSize >= page.TotalCount {
			return zones, nil
		}
	}
}

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ddnsx.Existing[Record], error) {
	const pageSize = 100

	rr, err := relativeRR(key.FQDN, key.Zone.Name.String())
	if err != nil {
		return nil, err
	}

	var records []ddnsx.Existing[Record]
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.DescribeDomainRecords(ctx, DescribeDomainRecordsRequest{
			DomainName:  key.Zone.Name.String(),
			RRKeyWord:   rr,
			TypeKeyWord: key.Type.String(),
			PageNumber:  pageNumber,
			PageSize:    pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeDomainRecords: %w", err)
		}
		for _, record := range page.DomainRecords.Record {
			if !strings.EqualFold(record.RR, rr) || record.Type != key.Type.String() {
				continue
			}
			address, err := netip.ParseAddr(record.Value)
			if err != nil {
				return nil, fmt.Errorf("record %s: not an address: %s: %w", record.RecordId, record.Value, err)
			}
			records = append(records, ddnsx.Existing[Record]{
				Addr:   address.Unmap(),
				Record: record,
			})
		}
		if len(page.DomainRecords.Record) < pageSize ||
			int64(pageNumber)*pageSize >= page.TotalCount {
			return records, nil
		}
	}
}

func (c *Client) Create(ctx context.Context, target ddnsx.RecordTarget) error {
	rr, err := relativeRR(target.FQDN, target.Zone.Name.String())
	if err != nil {
		return err
	}
	_, err = c.AddDomainRecord(ctx, AddDomainRecordRequest{
		DomainName: target.Zone.Name.String(),
		RR:         rr,
		Type:       target.Type.String(),
		Value:      target.Address.Unmap().String(),
		Line:       DefaultRecordLine,
		TTL:        target.TTL,
	})
	return err
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordTarget, record Record) error {
	rr, err := relativeRR(target.FQDN, target.Zone.Name.String())
	if err != nil {
		return err
	}
	_, err = c.UpdateDomainRecord(ctx, UpdateDomainRecordRequest{
		RecordId: record.RecordId,
		RR:       rr,
		Type:     target.Type.String(),
		Value:    target.Address.Unmap().String(),
		Line:     lineOrDefault(record.Line),
		TTL:      target.TTL,
	})
	return err
}

func (c *Client) Delete(ctx context.Context, _ ddnsx.RecordKey, record Record) error {
	_, err := c.DeleteDomainRecord(ctx, DeleteDomainRecordRequest{RecordId: record.RecordId})
	return err
}

func lineOrDefault(line string) string {
	if line == "" {
		return DefaultRecordLine
	}
	return line
}

func relativeRR(fqdn, zone string) (string, error) {
	fqdn, err := ddnsx.NormalizeFQDN(fqdn)
	if err != nil {
		return "", err
	}
	zone, err = ddnsx.NormalizeFQDN(zone)
	if err != nil {
		return "", err
	}
	if fqdn == zone {
		return ApexRecordHost, nil
	}
	if rr, found := strings.CutSuffix(fqdn, "."+zone); found {
		return rr, nil
	}
	return "", fmt.Errorf("fqdn %q is not within zone %q", fqdn, zone)
}

func (c *Client) actionLogger(action string) *zap.Logger {
	return c.logger.With(zap.String("action", action))
}

func doAction[Resp any](ctx context.Context, c *Client, action string, query urlpkg.Values) (Resp, error) {
	var zero Resp

	logger := c.actionLogger(action)

	req := newRequest(action)
	maps.Copy(req.Query, query)

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
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if jerr := json.NewDecoder(resp.Body).Decode(apiErr); jerr != nil {
			logger.Warn("decode error response failed",
				zap.Int("status", resp.StatusCode),
				zap.Error(jerr),
			)
			return zero, &httpx.BadStatusCodeError{Got: resp.StatusCode}
		}
		logger.Warn("api returned error",
			zap.String("error_code", apiErr.Code),
			zap.String("error_message", apiErr.Message),
			zap.String("request_id", apiErr.RequestID),
		)
		return zero, apiErr
	}

	var out Resp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logger.Warn("decode response failed",
			zap.Error(err),
		)

		return zero, fmt.Errorf("decode %s: %w", action, err)
	}
	return out, nil
}

func newRequest(action string) httpx.ReqConfig {
	r := httpx.NewReqConfig(http.MethodPost, AliyunDNSEndpoint)
	r.Query.Set(ParamAction, action)
	r.Query.Set(ParamVersion, AlidnsDefaultVersion)
	return r
}
