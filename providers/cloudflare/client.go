package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	urlpkg "net/url"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/providerx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

const (
	opDescribeDomains = providerx.OperationResolveZone
	opListRecords     = providerx.OperationListRecords
	opCreateRecord    = providerx.OperationCreateRecord
	opUpdateRecord    = providerx.OperationUpdateRecord
	opDeleteRecord    = providerx.OperationDeleteRecord
)

var (
	_ ddnsx.DDNSClient[DNSRecord] = (*Client)(nil)
	_ ddnsx.ZoneSearcher          = (*Client)(nil)
)

var cloudflareApiEndpoint = mt.Must(urlpkg.Parse("https://api.cloudflare.com/client/v4/zones"))

type Client struct {
	logger *zap.Logger
	do     httpx.HTTPRequester

	zones ddnsx.ZoneCache

	proxied        bool
	privateRouting bool
	comment        string
}

func NewClient(logger *zap.Logger, do httpx.HTTPRequester, token string,
	proxied, privateRouting bool, comment string,
) *Client {
	return &Client{
		logger: logger,
		do: &CloudflareHTTPRequester{
			HTTPRequester: do,
			Logger:        logger,
			Token:         token,
		},
		proxied:        proxied,
		privateRouting: privateRouting,
		comment:        comment,
	}
}

func (c *Client) ResolveZone(ctx context.Context, fqdn string) (ddnsx.Zone, error) {
	return c.zones.Resolve(ctx, fqdn, c)
}

func (c *Client) SearchZones(ctx context.Context, keyword ddnsx.ZoneName) ([]ddnsx.Zone, error) {
	logger := c.actionLogger(opDescribeDomains).With(zap.Stringer("keyword", keyword))
	logger.Info("search zone id from upstream")

	pager := c.ListZoneName(keyword.String())
	var zones []ddnsx.Zone
	for page, err := pager.Next(ctx); err != io.EOF; page, err = pager.Next(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list zones: %w", err)
		}
		for _, zone := range page {
			if !domains.IsDomainName(zone.Name) {
				logger.Warn("upstream returned a bad zone name", zap.String("zone_name", zone.Name))
				continue
			}
			zones = append(zones, ddnsx.Zone{
				Name: ddnsx.ZoneName(zone.Name),
				ID:   ddnsx.ZoneID(zone.ID),
			})
		}
	}
	return zones, nil
}

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ddnsx.Existing[DNSRecord], error) {
	pager, err := c.ListDNSRecords(key.FQDN, key.Zone.ID.String(), key.Type.String())
	if err != nil {
		return nil, fmt.Errorf("list DNS records: %w", err)
	}

	var existing []ddnsx.Existing[DNSRecord]
	for page, err := pager.Next(ctx); err != io.EOF; page, err = pager.Next(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list DNS records: %w", err)
		}
		for _, record := range page {
			address, err := netip.ParseAddr(record.Content)
			if err != nil {
				return nil, fmt.Errorf("record %s: not an address: %s: %w", record.ID, record.Content, err)
			}
			existing = append(existing, ddnsx.Existing[DNSRecord]{
				Addr:   address.Unmap(),
				Record: record,
			})
		}
	}
	return existing, nil
}

func (c *Client) Create(ctx context.Context, target ddnsx.RecordTarget) error {
	return c.CreateDNSRecords(ctx, target.Zone.ID.String(), c.recordRequest(target))
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordTarget, record DNSRecord) error {
	return c.UpdateDNSRecords(ctx, target.Zone.ID.String(), record.ID, c.recordRequest(target))
}

func (c *Client) Delete(ctx context.Context, key ddnsx.RecordKey, record DNSRecord) error {
	return c.DeleteDNSRecord(ctx, key.Zone.ID.String(), record.ID)
}

func (c *Client) recordRequest(target ddnsx.RecordTarget) UpdateDNSRecordRequest {
	return UpdateDNSRecordRequest{
		Name:           target.FQDN,
		Ttl:            target.TTL,
		Type:           target.Type.String(),
		Comment:        c.comment,
		Content:        target.Address.Unmap().String(),
		PrivateRouting: c.privateRouting,
		Proxied:        c.proxied,
	}
}

func (c *Client) NewRequestConfig(method string) httpx.ReqConfig {
	return httpx.NewReqConfig(method, cloudflareApiEndpoint)
}

func (c *Client) ListZones() *PageConfig[Zone] {
	return c.ListZoneName("")
}

func (c *Client) ListZoneName(name string) *PageConfig[Zone] {
	r := c.NewRequestConfig(http.MethodGet)
	r.Query.Set("status", "active")
	if name != "" {
		r.Query.Set("name", name)
	}
	return NewPaging[Zone](c, opDescribeDomains, r)
}

func (c *Client) ListDNSRecords(name string, zoneID string, qtype string) (*PageConfig[DNSRecord], error) {
	r := c.NewRequestConfig(http.MethodGet)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Query.Set("name", name)
	r.Query.Set("type", qtype)
	return NewPaging[DNSRecord](c, opListRecords, r), nil
}

func (c *Client) CreateDNSRecords(ctx context.Context, zoneID string,
	content UpdateDNSRecordRequest,
) (err error) {
	logger := c.actionLogger(opCreateRecord).With(zap.String("zone_id", zoneID))
	logger.Debug("api call start")
	r := c.NewRequestConfig(http.MethodPost)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Body = content
	createResult, response, perr := httpx.JSONRequest[Response](ctx, c.do, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := createResult.JoinError(perr); joined != nil {
		logger.Warn("create dns record failed", zap.Error(joined))
		err = joined
		return err
	}

	return nil
}

func (c *Client) UpdateDNSRecords(ctx context.Context, zoneID string, recordID string,
	content UpdateDNSRecordRequest,
) (err error) {
	logger := c.actionLogger(opUpdateRecord).With(
		zap.String("zone_id", zoneID),
		zap.String("record_id", recordID),
	)

	r := c.NewRequestConfig(http.MethodPatch)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", recordID)
	r.Body = content
	updateResult, response, perr := httpx.JSONRequest[Response](ctx, c.do, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := updateResult.JoinError(perr); joined != nil {
		logger.Warn("update dns record failed", zap.Error(joined))
		err = joined
		return err
	}
	return nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID string, dnsRecordID string) (err error) {
	logger := c.actionLogger(opDeleteRecord).With(
		zap.String("zone_id", zoneID),
		zap.String("record_id", dnsRecordID),
	)

	r := c.NewRequestConfig(http.MethodDelete)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", dnsRecordID)
	httpReq, perr := r.ToRequestContext(ctx)
	if perr != nil {
		err = fmt.Errorf("build request: %w", perr)
		return err
	}

	resp, perr := c.do.Do(httpReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if perr != nil {
		logger.Warn("delete request failed", zap.Error(perr))
		err = &httpx.BaseResponseError{Err: perr, Method: r.Method}
		return err
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		logger.Warn("delete bad status", zap.Int("status", resp.StatusCode))
		err = &httpx.BadStatusCodeError{Got: resp.StatusCode}
		return err
	}
	return nil
}

func (c *Client) actionLogger(action string) *zap.Logger {
	return c.logger.With(zap.String("action", action))
}
