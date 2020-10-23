package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// Deals provides access to Powergate deals APIs.
type Deals struct {
	client proto.PowergateServiceClient
}

// ListDealRecordsOption updates a ListDealRecordsConfig.
type ListDealRecordsOption func(*proto.ListDealRecordsConfig)

// WithFromAddrs limits the results deals initiated from the provided wallet addresses.
// If WithDataCids is also provided, this is an AND operation.
func WithFromAddrs(addrs ...string) ListDealRecordsOption {
	return func(c *proto.ListDealRecordsConfig) {
		c.FromAddrs = addrs
	}
}

// WithDataCids limits the results to deals for the provided data cids.
// If WithFromAddrs is also provided, this is an AND operation.
func WithDataCids(cids ...string) ListDealRecordsOption {
	return func(c *proto.ListDealRecordsConfig) {
		c.DataCids = cids
	}
}

// WithIncludePending specifies whether or not to include pending deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludePending(includePending bool) ListDealRecordsOption {
	return func(c *proto.ListDealRecordsConfig) {
		c.IncludePending = includePending
	}
}

// WithIncludeFinal specifies whether or not to include final deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludeFinal(includeFinal bool) ListDealRecordsOption {
	return func(c *proto.ListDealRecordsConfig) {
		c.IncludeFinal = includeFinal
	}
}

// WithAscending specifies to sort the results in ascending order. Default is descending order.
// Records are sorted by timestamp.
func WithAscending(ascending bool) ListDealRecordsOption {
	return func(c *proto.ListDealRecordsConfig) {
		c.Ascending = ascending
	}
}

// ListStorageDealRecords returns a list of storage deals for the FFS instance according to the provided options.
func (j *Jobs) ListStorageDealRecords(ctx context.Context, opts ...ListDealRecordsOption) (*proto.ListStorageDealRecordsResponse, error) {
	conf := &proto.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	return j.client.ListStorageDealRecords(ctx, &proto.ListStorageDealRecordsRequest{Config: conf})
}

// ListRetrievalDealRecords returns a list of retrieval deals for the FFS instance according to the provided options.
func (j *Jobs) ListRetrievalDealRecords(ctx context.Context, opts ...ListDealRecordsOption) (*proto.ListRetrievalDealRecordsResponse, error) {
	conf := &proto.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	return j.client.ListRetrievalDealRecords(ctx, &proto.ListRetrievalDealRecordsRequest{Config: conf})
}
