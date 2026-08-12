package tencentcloud

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/providerx"
	"github.com/duakc/lightddns/infra/netx/domains"
	"github.com/duakc/lightddns/infra/zaplog"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
)

var (
	_ APIClient                = (*defaultAPIClient)(nil)
	_ ddnsx.DDNSClient[Record] = (*Client)(nil)
	_ ddnsx.ZoneSearcher       = (*Client)(nil)
)

type Client struct {
	logger *zap.Logger
	api    APIClient
	zones  ddnsx.ZoneCache
}

func NewClient(logger *zap.Logger, api APIClient) *Client {
	return &Client{
		logger: zaplog.DoNotPanic(logger).Named("client"),
		api:    api,
	}
}

func (c *Client) ResolveZone(ctx context.Context, fqdn string) (ddnsx.Zone, error) {
	return c.zones.Resolve(ctx, fqdn, c)
}

func (c *Client) SearchZones(ctx context.Context, keyword string) ([]ddnsx.Zone, error) {
	const pageSize = 100

	// remove fqdn suffix dot.
	// tencent search operation doesn't support tail-dot.
	keyword = strings.TrimSuffix(keyword, ".")

	logger := c.logger.WithLazy(
		zap.String("action", "SearchZones"),
		zap.String("keyword", keyword),
	)
	logger.Info("search domain id from upstream")

	var zones []ddnsx.Zone
	for offset := 0; ; {
		page, err := c.api.DescribeDomainFilterList(ctx, DescribeDomainFilterListRequest{
			Type:    "ALL",
			Limit:   pageSize,
			Offset:  offset,
			Keyword: keyword,
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeDomainFilterList: %w", err)
		}
		for _, domain := range page.DomainList {
			zones = append(zones, ddnsx.Zone{
				Fqdn: mDns.Fqdn(domain.Name),
				ID:   strconv.FormatUint(domain.DomainId, 10),
			})
		}
		offset += len(page.DomainList)
		if len(page.DomainList) == 0 || offset >= page.DomainCountInfo.DomainTotal {
			return zones, nil
		}
	}
}

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ddnsx.Existing[Record], error) {
	const pageSize = 100

	domain := domains.FqdnToDomain(key.Zone.Fqdn)
	subdomain, err := relativeSubDomain(key.FQDN, key.Zone.Fqdn)
	if err != nil {
		return nil, err
	}

	var records []ddnsx.Existing[Record]
	for offset := 0; ; {
		page, err := c.api.DescribeRecordList(ctx, DescribeRecordListRequest{
			Domain:     domain,
			Subdomain:  subdomain,
			RecordType: key.Type.String(),
			Limit:      pageSize,
			Offset:     offset,
		})
		if err != nil {
			if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.Code == DNSPodErrCodeNoDataOfRecord {
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

func (c *Client) Create(ctx context.Context, target ddnsx.RecordSpec) error {
	domain := domains.FqdnToDomain(target.Zone.Fqdn)
	subdomain, err := relativeSubDomain(target.FQDN, target.Zone.Fqdn)
	if err != nil {
		return err
	}

	_, err = c.api.CreateRecord(ctx, CreateRecordRequest{
		Domain:     domain,
		SubDomain:  subdomain,
		RecordType: target.Type.String(),
		RecordLine: DefaultRecordLine,
		Value:      target.Address.Unmap().String(),
		TTL:        target.TTL,
		Remark:     providerx.UpdateMessage(""),
	})
	return err
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordSpec, record Record) error {
	domain := domains.FqdnToDomain(target.Zone.Fqdn)
	subdomain, err := relativeSubDomain(target.FQDN, target.Zone.Fqdn)
	if err != nil {
		return err
	}

	_, err = c.api.ModifyRecord(ctx, ModifyRecordRequest{
		Domain:     domain,
		RecordId:   record.RecordId,
		SubDomain:  subdomain,
		RecordType: target.Type.String(),
		RecordLine: cmp.Or(record.Line, DefaultRecordLine),
		Value:      target.Address.Unmap().String(),
		TTL:        target.TTL,
		Remark:     providerx.UpdateMessage(""),
	})
	return err
}

func (c *Client) Delete(ctx context.Context, key ddnsx.RecordKey, record Record) error {
	domain := domains.FqdnToDomain(key.Zone.Fqdn)
	_, err := c.api.DeleteRecord(ctx, DeleteRecordRequest{
		Domain:   domain,
		RecordId: record.RecordId,
	})
	return err
}

func relativeSubDomain(fqdn, zone string) (string, error) {
	relative, err := domains.CutDomainSuffix(fqdn, zone)
	if err != nil {
		return "", err
	}
	return cmp.Or(relative, "@"), nil
}
