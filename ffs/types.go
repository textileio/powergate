package ffs

import (
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
	EmptyID = InstanceID("")
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
	// and didn't reach executing.
	Cancelled
	// Done indicates the Job was successfully executed.
	Done
)

// CidConfigID is an identifier for a Cid configuration.
type CidConfigID string

// NewCidConfigID returns a new CidConfigID.
func NewCidConfigID() CidConfigID {
	return CidConfigID(uuid.New().String())
}

// Valid returns true if is a valid CidConfiID, false
// otherwise.
func (i CidConfigID) Valid() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}

// String returns a string representation of a CidConfigID.
func (i CidConfigID) String() string {
	return string(i)
}

// Job is a task to enforce CidConfig for a Cid.
type Job struct {
	ID       JobID
	Status   JobStatus
	ErrCause string
	Action   AddAction
	CidInfo  CidInfo
}

type CidConfig struct {
	Hot  HotConfig
	Cold ColdConfig
}

func (c CidConfig) WithColdFilEnabled(enabled bool) CidConfig {
	c.Cold.Filecoin.Enabled = enabled
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

func (c CidConfig) WithHotIpfsEnabled(enabled bool) CidConfig {
	c.Hot.Ipfs.Enabled = enabled
	return c
}

func (c CidConfig) WithHotIpfsAddTimeout(seconds int) CidConfig {
	c.Hot.Ipfs.AddTimeout = seconds
	return c
}

func (c CidConfig) Validate() error {
	if err := c.Hot.Ipfs.Validate(); err != nil {
		return fmt.Errorf("hot-ipfs config is invalid: %s", err)
	}
	if err := c.Cold.Filecoin.Validate(); err != nil {
		return fmt.Errorf("cold-filecoin config is invalid: %s", err)
	}
	return nil
}

// AddAction is the desired state of storage for a Cid.
type AddAction struct {
	ID         CidConfigID
	InstanceID InstanceID
	Cid        cid.Cid
	Config     CidConfig
	Meta       AddMeta
}

// AddMeta contains necessary metadata to execute the Add action.
type AddMeta struct {
	WalletAddr string
}

// Hotconfig is the desired storage of a Cid in a hot layer.
type HotConfig struct {
	Ipfs IpfsConfig
}

// IpfsConfig is the desired storage of a Cid in IPFS.
type IpfsConfig struct {
	Enabled    bool
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
	Filecoin FilecoinConfig
}

// FilecoinConfig is the desired state of a Cid in the
// Filecoin network.
type FilecoinConfig struct {
	Enabled      bool
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
	ConfigID CidConfigID
	Cid      cid.Cid
	Created  time.Time
	Hot      HotInfo
	Cold     ColdInfo
}

// HotInfo contains information about the current storage state
// of a Cid in the hot layer.
type HotInfo struct {
	Size int
	Ipfs IpfsHotInfo
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
