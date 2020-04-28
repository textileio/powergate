package ffs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
)

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
	// Queued indicates the Job is queued in the Scheduler.
	Queued JobStatus = iota
	// InProgress indicates that the Job is currently being
	// executed.
	InProgress
	// Failed indicates the Job failed, with job.ErrCause with
	// the error cause.
	Failed
	// Canceled indicates the Job was canceled from Queued,
	// and didn't reach execution.
	Canceled
	// Success indicates the Job was successfully executed.
	Success
)

// Job is a task executed by the Scheduler.
type Job struct {
	ID       JobID
	APIID    APIID
	Cid      cid.Cid
	Status   JobStatus
	ErrCause string
}

// DefaultConfig contains a default storage configuration for an Api instance.
type DefaultConfig struct {
	Hot  HotConfig
	Cold ColdConfig
}

// Validate validates a default Cid configuration.
func (dc DefaultConfig) Validate() error {
	if err := dc.Hot.Validate(); err != nil {
		return err
	}
	if err := dc.Cold.Validate(); err != nil {
		return err
	}
	return nil
}

// CidConfig has a Cid desired storing configuration for a Cid in
// Hot and Cold storage.
type CidConfig struct {
	// Cid is the Cid of the stored data.
	Cid cid.Cid
	// Hot has desired storing configuration in Hot Storage.
	Hot HotConfig
	// Cold has desired storing configuration in the Cold Storage.
	Cold ColdConfig
	// Repairable indicates if this cid config should be tracked
	// for auto-repair.
	Repairable bool
}

// WithRepairable allows to enable/disable auto-repair.
func (c CidConfig) WithRepairable(enabled bool) CidConfig {
	c.Repairable = enabled
	return c
}

// WithColdEnabled allows to enable/disable Cold storage usage.
func (c CidConfig) WithColdEnabled(enabled bool) CidConfig {
	c.Cold.Enabled = enabled
	return c
}

// WithColdFilCountryCodes defines a list of allowed country codes to select miners
// for deals.
func (c CidConfig) WithColdFilCountryCodes(countryCodes []string) CidConfig {
	c.Cold.Filecoin.CountryCodes = make([]string, len(countryCodes))
	copy(c.Cold.Filecoin.CountryCodes, countryCodes)
	return c
}

// WithColdFilExcludedMiners defines a list of miner addresses which won't be selected for
// making deals, no matter if they comply to other filters in the configuration.
func (c CidConfig) WithColdFilExcludedMiners(miners []string) CidConfig {
	c.Cold.Filecoin.ExcludedMiners = make([]string, len(miners))
	copy(c.Cold.Filecoin.ExcludedMiners, miners)
	return c
}

// WithColdFilRepFactor defines the replication factor for Filecoin storage.
func (c CidConfig) WithColdFilRepFactor(repFactor int) CidConfig {
	c.Cold.Filecoin.RepFactor = repFactor
	return c
}

// WithColdFilDealDuration defines the duration used for deals for Filecoin storage.
func (c CidConfig) WithColdFilDealDuration(duration int64) CidConfig {
	c.Cold.Filecoin.DealDuration = duration
	return c
}

// WithColdFilRenew specifies if deals should be renewed before they expire with a particular
// threshold chain epochs.
func (c CidConfig) WithColdFilRenew(enabled bool, threshold int) CidConfig {
	c.Cold.Filecoin.Renew.Enabled = enabled
	c.Cold.Filecoin.Renew.Threshold = threshold
	return c
}

// WithHotEnabled allows to enable/disable Hot storage usage.
func (c CidConfig) WithHotEnabled(enabled bool) CidConfig {
	c.Hot.Enabled = enabled
	return c
}

// WithHotIpfsAddTimeout specifies a timeout for fetching data in Ipfs.
func (c CidConfig) WithHotIpfsAddTimeout(seconds int) CidConfig {
	c.Hot.Ipfs.AddTimeout = seconds
	return c
}

// WithHotAllowUnfreeze allows the Scheduler to fetch data from the Cold Storage,
// if the Enabled flag of the Hot Storage switches from false->true.
func (c CidConfig) WithHotAllowUnfreeze(allow bool) CidConfig {
	c.Hot.AllowUnfreeze = true
	return c
}

// Validate validates a Cid configuration.
func (c CidConfig) Validate() error {
	if !c.Cid.Defined() {
		return fmt.Errorf("cid is undefined")
	}
	if err := c.Hot.Validate(); err != nil {
		return fmt.Errorf("hot-ipfs config is invalid: %s", err)
	}
	if err := c.Cold.Validate(); err != nil {
		return fmt.Errorf("cold-filecoin config is invalid: %s", err)
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
	// Ipfs contains configuration related to storing Cid data in a IPFS node.
	Ipfs IpfsConfig
}

// Validate validates a HotConfig.
func (hc HotConfig) Validate() error {
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
	if err := cc.Filecoin.Validate(); err != nil {
		return fmt.Errorf("invalid Filecoin config: %s", err)
	}
	if cc.Enabled && cc.Filecoin.Addr == "" {
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
	// DealDuration indicates the duration to be used when making new deals.
	DealDuration int64
	// ExcludedMiners is a set of miner addresses won't be ever be selected
	// when making new deals, even if they comply to other filters.
	ExcludedMiners []string
	// CountryCodes indicates that new deals should select miners on specific
	// countries.
	CountryCodes []string
	// FilRenew indicates deal-renewal configuration.
	Renew FilRenew
	// Addr is the wallet address used to store the data in filecoin
	Addr string
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
func (fc *FilConfig) Validate() error {
	if fc.RepFactor <= 0 {
		return fmt.Errorf("replication factor should be greater than zero, got %d", fc.RepFactor)
	}
	if fc.DealDuration <= 0 {
		return fmt.Errorf("deal duration should be greater than zero, got %d", fc.DealDuration)
	}
	if fc.Renew.Enabled && fc.Renew.Threshold <= 0 {
		return fmt.Errorf("renew threshold should be positive: %d", fc.Renew.Threshold)
	}
	return nil
}

// CidInfo contains information about the current storage state
// of a Cid.
type CidInfo struct {
	JobID   JobID
	Cid     cid.Cid
	Created time.Time
	Hot     HotInfo
	Cold    ColdInfo
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
	Filecoin FilInfo
}

// FilInfo contains information about the current storage state
// of a Cid in the Filecoin network.
type FilInfo struct {
	DataCid   cid.Cid
	Proposals []FilStorage
}

// FilStorage contains Deal information of a storage in Filecoin.
type FilStorage struct {
	ProposalCid     cid.Cid
	Renewed         bool
	Duration        int64
	ActivationEpoch int64
	Miner           string
	Size            uint64
}

// CidLoggerCtxKey is a type to use in ctx values for CidLogger.
type CidLoggerCtxKey int

const (
	// CtxKeyJid is the key to store Jid metadata.
	CtxKeyJid CidLoggerCtxKey = iota
)

// CidLogger saves log information about a Cid executions.
type CidLogger interface {
	Log(context.Context, cid.Cid, string, ...interface{})
	Watch(context.Context, chan<- LogEntry) error
}

// LogEntry is a log entry from a Cid execution.
type LogEntry struct {
	Cid       cid.Cid
	Timestamp time.Time
	Jid       JobID
	Msg       string
}
