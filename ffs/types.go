package ffs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/util"
)

var (
	// EmptyRetrievalID is an undef retrieval id.
	EmptyRetrievalID = RetrievalID("")
)

// RetrievalID is the identifier of a Filecoin retrieval.
type RetrievalID string

// NewRetrievalID returns an new retrieval ID.
func NewRetrievalID() RetrievalID {
	return RetrievalID(uuid.New().String())
}

// String returns a string representation of RetrievalID.
func (ri RetrievalID) String() string {
	return string(ri)
}

var (
	// EmptyJobID represents an empty JobID.
	EmptyJobID = JobID("")
)

// JobID is an identifier for a ffs.Job.
type JobID string

// NewJobID returns a new JobID.
func NewJobID() JobID {
	return JobID(uuid.New().String())
}

// String returns a string representation of JobID.
func (jid JobID) String() string {
	return string(jid)
}

var (
	// EmptyInstanceID represents an empty/invalid Instance ID.
	EmptyInstanceID = APIID("")
)

// APIID is an identifier for a Api instance.
type APIID string

// NewAPIID returns a new InstanceID.
func NewAPIID() APIID {
	return APIID(uuid.New().String())
}

// Valid returns true if the InstanceID is valid, false
// otherwise.
func (i APIID) Valid() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}

// String returns a string representation of InstanceID.
func (i APIID) String() string {
	return string(i)
}

// JobStatus is a type for Job statuses.
type JobStatus int

const (
	// Unspecified indicates a default or empty value
	Unspecified JobStatus = iota
	// Queued indicates the Job is queued in the Scheduler.
	Queued
	// Executing indicates that the Job is currently being
	// executed.
	Executing
	// Failed indicates the Job failed, with job.ErrCause with
	// the error cause.
	Failed
	// Canceled indicates the Job was canceled from Queued,
	// and didn't reach execution.
	Canceled
	// Success indicates the Job was successfully executed.
	Success
)

// JobStatusStr maps JobStatus to describing string.
var JobStatusStr = map[JobStatus]string{
	Unspecified: "Unspecified",
	Queued:      "Queued",
	Executing:   "Executing",
	Failed:      "Failed",
	Canceled:    "Canceled",
	Success:     "Success",
}

// StorageJob is a task executed by the Scheduler.
type StorageJob struct {
	ID         JobID
	APIID      APIID
	Cid        cid.Cid
	Status     JobStatus
	ErrCause   string
	DealErrors []DealError
	CreatedAt  int64
}

// RetrievalJob is a retrieval task executed by the Scheduler.
type RetrievalJob struct {
	ID          JobID
	APIID       APIID
	RetrievalID RetrievalID
	Status      JobStatus
	ErrCause    string
}

// StorageConfig contains a default storage configuration for an Api instance.
type StorageConfig struct {
	Hot        HotConfig
	Cold       ColdConfig
	Repairable bool
}

// WithRepairable allows to enable/disable auto-repair.
func (s StorageConfig) WithRepairable(enabled bool) StorageConfig {
	s.Repairable = enabled
	return s
}

// WithColdEnabled allows to enable/disable Cold storage usage.
func (s StorageConfig) WithColdEnabled(enabled bool) StorageConfig {
	s.Cold.Enabled = enabled
	return s
}

// WithColdFastRetrieval sets the Fast Retrieval feature for new deals.
func (s StorageConfig) WithColdFastRetrieval(enabled bool) StorageConfig {
	s.Cold.Filecoin.FastRetrieval = enabled
	return s
}

// WithColdStartDealOffset sets the maximum time in epochs a new deal must
// be active on-chain.
func (s StorageConfig) WithColdStartDealOffset(offset int64) StorageConfig {
	s.Cold.Filecoin.DealStartOffset = offset
	return s
}

// WithColdFilCountryCodes defines a list of allowed country codes to select miners
// for deals.
func (s StorageConfig) WithColdFilCountryCodes(countryCodes []string) StorageConfig {
	s.Cold.Filecoin.CountryCodes = make([]string, len(countryCodes))
	copy(s.Cold.Filecoin.CountryCodes, countryCodes)
	return s
}

// WithColdFilExcludedMiners defines a list of miner addresses which won't be selected for
// making deals, no matter if they comply to other filters in the configuration.
func (s StorageConfig) WithColdFilExcludedMiners(miners []string) StorageConfig {
	s.Cold.Filecoin.ExcludedMiners = make([]string, len(miners))
	copy(s.Cold.Filecoin.ExcludedMiners, miners)
	return s
}

// WithColdFilTrustedMiners defines a list of trusted miners addresses which will be
// returned if available. If more miners reusults are needed, other filters will be
// applied as usual.
func (s StorageConfig) WithColdFilTrustedMiners(miners []string) StorageConfig {
	s.Cold.Filecoin.TrustedMiners = make([]string, len(miners))
	copy(s.Cold.Filecoin.TrustedMiners, miners)
	return s
}

// WithColdFilRepFactor defines the replication factor for Filecoin storage.
func (s StorageConfig) WithColdFilRepFactor(repFactor int) StorageConfig {
	s.Cold.Filecoin.RepFactor = repFactor
	return s
}

// WithColdFilDealDuration defines the duration used for deals for Filecoin storage.
func (s StorageConfig) WithColdFilDealDuration(duration int64) StorageConfig {
	s.Cold.Filecoin.DealMinDuration = duration
	return s
}

// WithColdFilRenew specifies if deals should be renewed before they expire with a particular
// threshold chain epochs.
func (s StorageConfig) WithColdFilRenew(enabled bool, threshold int) StorageConfig {
	s.Cold.Filecoin.Renew.Enabled = enabled
	s.Cold.Filecoin.Renew.Threshold = threshold
	return s
}

// WithColdMaxPrice specifies the max price that should be considered for
// deal asks even when all other filers match.
func (s StorageConfig) WithColdMaxPrice(maxPrice uint64) StorageConfig {
	s.Cold.Filecoin.MaxPrice = maxPrice
	return s
}

// WithFastRetrieval specifies if deal fast retrieval flag on new deals
// is enabled.
func (s StorageConfig) WithFastRetrieval(enabled bool) StorageConfig {
	s.Cold.Filecoin.FastRetrieval = enabled
	return s
}

// WithColdAddr specifies the wallet address that should be used for transactions.
func (s StorageConfig) WithColdAddr(addr string) StorageConfig {
	s.Cold.Filecoin.Addr = addr
	return s
}

// WithHotEnabled allows to enable/disable Hot storage usage.
func (s StorageConfig) WithHotEnabled(enabled bool) StorageConfig {
	s.Hot.Enabled = enabled
	return s
}

// WithHotIpfsAddTimeout specifies a timeout for fetching data in Ipfs.
func (s StorageConfig) WithHotIpfsAddTimeout(seconds int) StorageConfig {
	s.Hot.Ipfs.AddTimeout = seconds
	return s
}

// WithHotAllowUnfreeze allows the Scheduler to fetch data from the Cold Storage,
// if the Enabled flag of the Hot Storage switches from false->true.
func (s StorageConfig) WithHotAllowUnfreeze(allow bool) StorageConfig {
	s.Hot.AllowUnfreeze = true
	return s
}

// WithUnfreezeMaxPrice indicates the maximum price to pay for an unfreeze
// procedure.
func (s StorageConfig) WithUnfreezeMaxPrice(maxPrice uint64) StorageConfig {
	s.Hot.UnfreezeMaxPrice = maxPrice
	return s
}

// Validate validates a StorageConfig.
func (s StorageConfig) Validate() error {
	if err := s.Hot.Validate(); err != nil {
		return fmt.Errorf("hot-ipfs config is invalid: %s", err)
	}
	if err := s.Cold.Validate(); err != nil {
		return fmt.Errorf("cold-filecoin config is invalid: %s", err)
	}
	// We can't accept being renewable without the hot storage enabled.
	// See the (**) note in scheduler.go
	if s.Cold.Enabled && s.Cold.Filecoin.Renew.Enabled && !s.Hot.Enabled {
		return fmt.Errorf("hot storage should be enabled to enable renewals")
	}
	return nil
}

// HotConfig is the desired storage of a Cid in a Hot Storage.
type HotConfig struct {
	// Enable indicates if Cid data is stored. If true, it will consider
	// further configurations to execute actions.
	Enabled bool
	// AllowUnfreeze indicates that if data isn't available in the Hot Storage,
	// it's allowed to be feeded by Cold Storage if available.
	AllowUnfreeze bool
	// UnfreezeMaxPrice indicates the maximum amount of attoFil to pay for
	// retrieval of data to unfreeze.
	UnfreezeMaxPrice uint64
	// Ipfs contains configuration related to storing Cid data in a IPFS node.
	Ipfs IpfsConfig
}

// Validate validates a HotConfig.
func (hc HotConfig) Validate() error {
	if !hc.Enabled {
		return nil
	}
	if err := hc.Ipfs.Validate(); err != nil {
		return fmt.Errorf("invalid ipfs config: %s", err)
	}
	return nil
}

// IpfsConfig is the desired storage of a Cid in IPFS.
type IpfsConfig struct {
	// AddTimeout is an upper bound on adding data to IPFS node from
	// the network before failing.
	AddTimeout int
}

// Validate validates an IpfsConfig.
func (ic *IpfsConfig) Validate() error {
	if ic.AddTimeout <= 0 {
		return fmt.Errorf("add timeout should be greater than 0 seconds, got %d", ic.AddTimeout)
	}
	return nil
}

// ColdConfig is the desired state of a Cid in a cold layer.
type ColdConfig struct {
	// Enabled indicates that data will be saved in Cold storage.
	// If is switched from false->true, it will consider the other attributes
	// as the desired state of the data in this Storage.
	Enabled bool
	// Filecoin describes the desired Filecoin configuration for a Cid in the
	// Filecoin network.
	Filecoin FilConfig
}

// Validate validates a ColdConfig.
func (cc ColdConfig) Validate() error {
	if !cc.Enabled {
		return nil
	}
	if err := cc.Filecoin.Validate(); err != nil {
		return fmt.Errorf("invalid Filecoin config: %s", err)
	}
	if cc.Filecoin.Addr == "" {
		return fmt.Errorf("invalid wallet address")
	}
	return nil
}

// FilConfig is the desired state of a Cid in the Filecoin network.
type FilConfig struct {
	// RepFactor indicates the desired amount of active deals
	// with different miners to store the data. While making deals
	// the other attributes of FilConfig are considered for miner selection.
	RepFactor int
	// DealMinDuration indicates the duration to be used when making new deals.
	DealMinDuration int64
	// ExcludedMiners is a set of miner addresses won't be ever be selected
	// when making new deals, even if they comply to other filters.
	ExcludedMiners []string
	// TrustedMiners is a set of miner addresses which will be forcibly used
	// when making new deals. An empty/nil list disables this feature.
	TrustedMiners []string
	// CountryCodes indicates that new deals should select miners on specific
	// countries.
	CountryCodes []string
	// Renew indicates deal-renewal configuration.
	Renew FilRenew
	// Addr is the wallet address used to store the data in filecoin
	Addr string
	// MaxPrice is the maximum price that will be spent to store the data
	MaxPrice uint64
	// FastRetrieval indicates that created deals should enable the
	// fast retrieval feature.
	FastRetrieval bool
	// DealStartOffset indicates how many epochs in the future impose a
	// deadline to new deals being active on-chain. This value might influence
	// if miners accept deals, since they should seal fast enough to satisfy
	// this constraint.
	DealStartOffset int64
}

// Validate returns a non-nil error if the configuration is invalid.
func (fc *FilConfig) Validate() error {
	if fc.RepFactor <= 0 {
		return fmt.Errorf("replication factor should be greater than zero, got %d", fc.RepFactor)
	}
	if fc.DealMinDuration < util.MinDealDuration {
		return fmt.Errorf("deal duration should be greater than minimum, got %d", fc.DealMinDuration)
	}
	if err := fc.Renew.Validate(); err != nil {
		return fmt.Errorf("invalid renew config: %s", err)
	}
	return nil
}

// FilRenew contains renew configuration for a Cid Cold Storage deals.
type FilRenew struct {
	// Enabled indicates that deal-renewal is enabled for this Cid.
	Enabled bool
	// Threshold indicates how many epochs before expiring should trigger
	// deal renewal. e.g: 100 epoch before expiring.
	Threshold int
}

// Validate returns a non-nil error if the configuration is invalid.
func (fr *FilRenew) Validate() error {
	if fr.Enabled && fr.Threshold <= 0 {
		return fmt.Errorf("renew threshold should be positive: %d", fr.Threshold)
	}
	return nil
}

// RetrievalInfo has data about an executed Filecoin retrieval.
type RetrievalInfo struct {
	ID        RetrievalID
	DataCid   cid.Cid
	TotalPaid uint64
	MinerAddr string
	Size      int64
	CreatedAt time.Time
}

// CidInfo contains information about the current storage state
// of a Cid.
type CidInfo struct {
	// JobID indicates the Job ID which updated
	// the current information. It *may be empty* if
	// the data was imported manually.
	JobID JobID
	// Cid of payload.
	Cid cid.Cid
	// Created is the timestamp of the data.
	Created time.Time
	// Hot contains hot storage information.
	Hot HotInfo
	// Cold contains cold storage information.
	Cold ColdInfo
}

// HotInfo contains information about the current storage state
// of a Cid in the hot layer.
type HotInfo struct {
	Enabled bool
	Size    int
	Ipfs    IpfsHotInfo
}

// IpfsHotInfo contains information about the current storage state
// of a Cid in an IPFS node.
type IpfsHotInfo struct {
	Created time.Time
}

// ColdInfo contains information about the current storage state
// of a Cid in the cold layer.
type ColdInfo struct {
	Enabled  bool
	Filecoin FilInfo
}

// FilInfo contains information about the current storage state
// of a Cid in the Filecoin network.
type FilInfo struct {
	// DataCid corresponds to the PayloadCid of the deal.
	DataCid cid.Cid
	// Size is the size of the Piece. Recall that this size
	// is which is accounted for payment. Also is equal or
	// greater than the original data size.
	// This value might be zero for imported deals; if that's
	// the case, will be re-calculated in the next made deal.
	Size uint64
	// Proposals contains known deals for the data.
	Proposals []FilStorage
}

// FilStorage contains Deal information of a storage in Filecoin.
// This information is used in FFS may be used by FFS logic to
// provide repair, renwal, or retrieval tasks.
type FilStorage struct {
	// ProposalCid of the deal.
	ProposalCid cid.Cid
	// PieceCid is the piece Cid.
	PieceCid cid.Cid
	// Renewed indicates if this deal was
	// already renewed, so it can expiry
	// safely if renewals are enabled.
	Renewed bool
	// Duration is the duration of the deal.
	Duration int64
	// ActivationEpoch is the epoch in which
	// the deal was activated.
	ActivationEpoch int64
	// StartEpoch is the starting epoch in which
	// the deal is considered active on-chain.
	StartEpoch uint64
	// Miner is the miner address which is storing
	// deals data.
	Miner string
	// EpochPrice is the price of attoFil per GiB
	// per epoch paid in this deal.
	EpochPrice uint64
}

// JobLoggerCtxKey is a type to use in ctx values for CidLogger.
type JobLoggerCtxKey int

const (
	// CtxKeyJid is a context-key to indicate the Job ID for JobLogger.
	CtxKeyJid JobLoggerCtxKey = iota
	// CtxStorageCid is the context-key to indicate the Cid of a
	// StorageJob for JobLogger.
	CtxStorageCid
	// CtxRetrievalID is the context-key to indicate the RetrievalID of
	// a RetrievalJob for JobLogger.
	CtxRetrievalID
)

// JobLogger saves log information about a storage and retrieval tasks.
type JobLogger interface {
	Log(context.Context, string, ...interface{})
	Watch(context.Context, chan<- LogEntry) error
	GetByCid(context.Context, cid.Cid) ([]LogEntry, error)
}

// LogEntry is a log entry from a Cid execution.
type LogEntry struct {
	Cid       cid.Cid
	Timestamp time.Time
	Jid       JobID
	Msg       string
}

// PaychDir specifies the direction of a payment channel.
type PaychDir int

const (
	// PaychDirUnspecified is an undefined direction.
	PaychDirUnspecified PaychDir = iota
	// PaychDirInbound is an inbound direction.
	PaychDirInbound
	// PaychDirOutbound is an outbound direction.
	PaychDirOutbound
)

// PaychDirStr maps PaychDirs to describing string.
var PaychDirStr = map[PaychDir]string{
	PaychDirUnspecified: "Unspecified",
	PaychDirInbound:     "Inbound",
	PaychDirOutbound:    "Outbound",
}

// PaychInfo holds information about a payment channel.
type PaychInfo struct {
	CtlAddr   string
	Addr      string
	Direction PaychDir
}
