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
	// EmptyInstanceID representes an empty/invalid Instance ID.
	EmptyInstanceID = ApiID("")
)

// ApiID is an identifier for a Api instance.
type ApiID string

// NewApiID returns a new InstanceID.
func NewApiID() ApiID {
	return ApiID(uuid.New().String())
}

// Valid returns true if the InstanceID is valid, false
// otherwise.
func (i ApiID) Valid() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}

// String returns a string representation of InstanceID.
func (i ApiID) String() string {
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
	// Cancelled indicates the Job was cancelled from Queued,
	// and didn't reach execution.
	Cancelled
	// Success indicates the Job was successfully executed.
	Success
)

// Job is a task executed by the Scheduler.
type Job struct {
	ID         JobID
	InstanceID ApiID
	Status     JobStatus
	ErrCause   string
}

// DefaultCidConfig contains a default Cid configuration for an Api.
type DefaultCidConfig struct {
	Hot  HotConfig
	Cold ColdConfig
}

// Validate validates a default Cid configuration.
func (dc DefaultCidConfig) Validate() error {
	if err := dc.Hot.Validate(); err != nil {
		return err
	}
	if err := dc.Cold.Validate(); err != nil {
		return err
	}
	return nil
}

// CidConfig has a Cid desired storing configuration.
type CidConfig struct {
	Cid  cid.Cid
	Hot  HotConfig
	Cold ColdConfig
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
	if !c.Hot.Enabled && !c.Cold.Enabled {
		return ErrBothStoragesDisabled
	}
	return nil
}

// PushConfigAction contains information for pushing a new Cid configuration to the Scheduler.
type PushConfigAction struct {
	InstanceID ApiID
	Config     CidConfig
	WalletAddr string
}

// Validate validates a a PushConfigAction.
func (aa PushConfigAction) Validate() error {
	if aa.InstanceID == EmptyInstanceID {
		return fmt.Errorf("invalid Action ID")
	}
	if err := aa.Config.Validate(); err != nil {
		return err
	}
	if aa.WalletAddr == "" {
		return fmt.Errorf("invalid wallet address")
	}
	return nil
}

// HotConfig is the desired storage of a Cid in a Hot Storage.
type HotConfig struct {
	Enabled       bool
	AllowUnfreeze bool
	Ipfs          IpfsConfig
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
	Enabled  bool
	Filecoin FilConfig
}

// Validate validates a ColdConfig.
func (cc ColdConfig) Validate() error {
	if err := cc.Filecoin.Validate(); err != nil {
		return fmt.Errorf("invalid Filecoin config: %s", err)
	}
	return nil
}

// FilConfig is the desired state of a Cid in the
// Filecoin network.
type FilConfig struct {
	RepFactor      int
	DealDuration   int64
	ExcludedMiners []string
	CountryCodes   []string
	Renew          FilRenew
}

// FilRenew contains renew configuration for a Cid Cold Storage deals.
type FilRenew struct {
	Enabled   bool
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
}

type CidLoggerCtxKey int

const (
	CtxKeyJid CidLoggerCtxKey = iota
)

type CidLogger interface {
	Log(context.Context, cid.Cid, string, ...interface{})
}
