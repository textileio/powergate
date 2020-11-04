package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// Deals provides access to Powergate deals APIs.
type Deals struct {
	client proto.PowergateServiceClient
}

// DealRecordsOption updates a ListDealRecordsConfig.
type DealRecordsOption func(*proto.DealRecordsConfig)

// WithFromAddrs limits the results deals initiated from the provided wallet addresses.
// If WithDataCids is also provided, this is an AND operation.
func WithFromAddrs(addrs ...string) DealRecordsOption {
	return func(c *proto.DealRecordsConfig) {
		c.FromAddrs = addrs
	}
}

// WithDataCids limits the results to deals for the provided data cids.
// If WithFromAddrs is also provided, this is an AND operation.
func WithDataCids(cids ...string) DealRecordsOption {
	return func(c *proto.DealRecordsConfig) {
		c.DataCids = cids
	}
}

// WithIncludePending specifies whether or not to include pending deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludePending(includePending bool) DealRecordsOption {
	return func(c *proto.DealRecordsConfig) {
		c.IncludePending = includePending
	}
}

// WithIncludeFinal specifies whether or not to include final deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludeFinal(includeFinal bool) DealRecordsOption {
	return func(c *proto.DealRecordsConfig) {
		c.IncludeFinal = includeFinal
	}
}

// WithAscending specifies to sort the results in ascending order. Default is descending order.
// Records are sorted by timestamp.
func WithAscending(ascending bool) DealRecordsOption {
	return func(c *proto.DealRecordsConfig) {
		c.Ascending = ascending
	}
}

// StorageDealRecords returns a list of storage deals for the storage profile according to the provided options.
func (d *Deals) StorageDealRecords(ctx context.Context, opts ...DealRecordsOption) (*proto.StorageDealRecordsResponse, error) {
	conf := &proto.DealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	return d.client.StorageDealRecords(ctx, &proto.StorageDealRecordsRequest{Config: conf})
}

// RetrievalDealRecords returns a list of retrieval deals for the storage profile according to the provided options.
func (d *Deals) RetrievalDealRecords(ctx context.Context, opts ...DealRecordsOption) (*proto.RetrievalDealRecordsResponse, error) {
	conf := &proto.DealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	return d.client.RetrievalDealRecords(ctx, &proto.RetrievalDealRecordsRequest{Config: conf})
}
