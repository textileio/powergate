package ffs

import (
	"encoding/json"
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
	enabled    bool
	addTimeout int
}

func NewIpfsConfig(enabled bool, addTimeout int) (IpfsConfig, error) {
	if addTimeout <= 0 {
		return IpfsConfig{}, fmt.Errorf("time duration in seconds should be positive")
	}
	return IpfsConfig{
		enabled:    enabled,
		addTimeout: addTimeout,
	}, nil
}

func (ic IpfsConfig) Enabled() bool {
	return ic.enabled
}

func (ic IpfsConfig) AddTimeout() int {
	return ic.addTimeout
}

func (ic IpfsConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Enabled    bool
		AddTimeout int
	}{
		Enabled:    ic.enabled,
		AddTimeout: ic.addTimeout,
	})
}

func (ic *IpfsConfig) UnmarshalJSON(b []byte) error {
	var aic struct {
		Enabled    bool
		AddTimeout int
	}
	if err := json.Unmarshal(b, &aic); err != nil {
		return err
	}
	_, err := NewIpfsConfig(aic.Enabled, aic.AddTimeout)
	if err != nil {
		return err
	}
	ic.enabled = aic.Enabled
	ic.addTimeout = aic.AddTimeout
	return nil
}

// ColdConfig is the desired state of a Cid in a cold layer.
type ColdConfig struct {
	Filecoin FilecoinConfig
}

// FilecoinConfig is the desired state of a Cid in the
// Filecoin network.
type FilecoinConfig struct {
	enabled   bool
	repFactor int
}

func NewFilecoinConfig(enabled bool, repFactor int) (FilecoinConfig, error) {
	if repFactor < 1 {
		return FilecoinConfig{}, fmt.Errorf("invalid rep factor")
	}
	return FilecoinConfig{
		enabled:   enabled,
		repFactor: repFactor,
	}, nil
}

func (fc FilecoinConfig) Enabled() bool {
	return fc.enabled
}

func (fc FilecoinConfig) RepFactor() int {
	return fc.repFactor
}

func (fc FilecoinConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Enabled   bool
		RepFactor int
	}{
		Enabled:   fc.enabled,
		RepFactor: fc.repFactor,
	})
}

func (fc *FilecoinConfig) UnmarshalJSON(b []byte) error {
	var afc struct {
		Enabled   bool
		RepFactor int
	}
	if err := json.Unmarshal(b, &afc); err != nil {
		return err
	}
	_, err := NewFilecoinConfig(afc.Enabled, afc.RepFactor)
	if err != nil {
		return err
	}
	fc.enabled = afc.Enabled
	fc.repFactor = afc.RepFactor
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
	Duration   uint64
	Proposals  []FilStorage
}

// FilStorage contains Deal information of a storage in Filecoin.
type FilStorage struct {
	ProposalCid cid.Cid
	Failed      bool
}
