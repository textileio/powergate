package ffs

import (
	"errors"
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

// InstanceID
var (
	// EmptyInstanceID representes an empty/invalid Instance ID.
	EmptyInstanceID = InstanceID("")
	// ErrBothStorageDisabled returned when both storages are disabled.
	ErrBothStoragesDisabled = errors.New("both Hot and Cold layers can't be disabled")
	// ErrHotStorageDisabled returned when trying to fetch a Cid when disabled on Hot Storage.
	// To retrieve the data, is necessary to call Unfreeze().
	ErrHotStorageDisabled = errors.New("cid disabled in hot storage")
)

// InstanceID is an identifier for a Api instance.
type InstanceID string

// NewInstanceID returns a new InstanceID.
func NewInstanceID() InstanceID {
	return InstanceID(uuid.New().String())
}

// Valid returns true if the InstanceID is valid, false
// otherwise.
func (i InstanceID) Valid() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}

// String returns a string representation of InstanceID.
func (i InstanceID) String() string {
	return string(i)
}

// JobStatus
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
	// Done indicates the Job was successfully executed.
	Success
)

// Job is a task executed by the Scheduler.
type Job struct {
	ID         JobID
	InstanceID InstanceID
	Status     JobStatus
	ErrCause   string
}

type DefaultCidConfig struct {
	Hot  HotConfig
	Cold ColdConfig
}

func (dc DefaultCidConfig) Validate() error {
	if err := dc.Hot.Validate(); err != nil {
		return err
	}
	if err := dc.Cold.Validate(); err != nil {
		return err
	}
	return nil
}

type CidConfig struct {
	Cid  cid.Cid
	Hot  HotConfig
	Cold ColdConfig
}

func (c CidConfig) WithColdEnabled(enabled bool) CidConfig {
	c.Cold.Enabled = enabled
	return c
}

func (c CidConfig) WithColdFilCountryCodes(countryCodes []string) CidConfig {
	c.Cold.Filecoin.CountryCodes = make([]string, len(countryCodes))
	copy(c.Cold.Filecoin.CountryCodes, countryCodes)
	return c
}

func (c CidConfig) WithColdFilBlacklist(blacklist []string) CidConfig {
	c.Cold.Filecoin.Blacklist = make([]string, len(blacklist))
	copy(c.Cold.Filecoin.Blacklist, blacklist)
	return c
}

func (c CidConfig) WithColdFilRepFactor(repFactor int) CidConfig {
	c.Cold.Filecoin.RepFactor = repFactor
	return c
}

func (c CidConfig) WithColdFilDealDuration(duration int64) CidConfig {
	c.Cold.Filecoin.DealDuration = duration
	return c
}

func (c CidConfig) WithHotEnabled(enabled bool) CidConfig {
	c.Hot.Enabled = enabled
	return c
}

func (c CidConfig) WithHotIpfsAddTimeout(seconds int) CidConfig {
	c.Hot.Ipfs.AddTimeout = seconds
	return c
}

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

type PushConfigAction struct {
	InstanceID InstanceID
	Config     CidConfig
	WalletAddr string
}

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

// Hotconfig is the desired storage of a Cid in a hot layer.
type HotConfig struct {
	Enabled       bool
	AllowUnfreeze bool
	Ipfs          IpfsConfig
}

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

func (ic *IpfsConfig) Validate() error {
	if ic.AddTimeout <= 0 {
		return fmt.Errorf("add timeout should be greater than 0 seconds, got %d", ic.AddTimeout)
	}
	return nil
}

// ColdConfig is the desired state of a Cid in a cold layer.
type ColdConfig struct {
	Enabled  bool
	Filecoin FilecoinConfig
}

func (cc ColdConfig) Validate() error {
	if err := cc.Filecoin.Validate(); err != nil {
		return fmt.Errorf("invalid Filecoin config: %s", err)
	}
	return nil
}

// FilecoinConfig is the desired state of a Cid in the
// Filecoin network.
type FilecoinConfig struct {
	RepFactor    int
	DealDuration int64
	Blacklist    []string
	CountryCodes []string
}

// Validate returns a non-nil error if the configuration is invalid.
func (fc *FilecoinConfig) Validate() error {
	if fc.RepFactor <= 0 {
		return fmt.Errorf("replication factor should be greater than zero, got %d", fc.RepFactor)
	}
	if fc.DealDuration <= 0 {
		return fmt.Errorf("deal duration should be greater than zero, got %d", fc.DealDuration)
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
	Enabled  bool
	Filecoin FilInfo
}

// FilInfo contains information about the current storage state
// of a Cid in the Filecoin network.
type FilInfo struct {
	PayloadCID cid.Cid
	Proposals  []FilStorage
}

// FilStorage contains Deal information of a storage in Filecoin.
type FilStorage struct {
	ProposalCid     cid.Cid
	Duration        int64
	ActivationEpoch uint64
	Failed          bool
	Miner           string
}
