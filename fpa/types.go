package fpa

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
)

type WalletManager interface {
	NewWallet(ctx context.Context, typ string) (string, error)
	Balance(ctx context.Context, addr string) (uint64, error)
}

type Auditor interface {
	Start(ctx context.Context, instanceID string) OpAuditor
}
type OpAuditor interface {
	ID() string
	Success()
	Errored(error)
	Log(event interface{})
	Close()
}

type Scheduler interface {
	Enqueue(CidConfig) (JobID, error)

	Watch(InstanceID) <-chan Job
	Unwatch(<-chan Job)
}

type HotLayer interface {
	Add(context.Context, io.Reader) (cid.Cid, error)
	Get(context.Context, cid.Cid) (io.Reader, error)
	Pin(context.Context, cid.Cid) (HotInfo, error)
}

type ColdLayer interface {
	Store(ctx context.Context, c cid.Cid, conf ColdConfig) (ColdInfo, error)
}

var (
	EmptyJobID = JobID("")
)

type JobID string

func NewJobID() JobID {
	return JobID(uuid.New().String())
}

func (jid JobID) String() string {
	return string(jid)
}

type JobStatus int

const (
	Queued JobStatus = iota
	Failed
	Cancelled
	Done
)

type Job struct {
	ID       JobID
	Status   JobStatus
	ErrCause string
	Config   CidConfig
	CidInfo  CidInfo
}

type Config struct {
	Hot  HotConfig
	Cold ColdConfig
}

type HotConfig struct {
	Ipfs IpfsConfig
}

type IpfsConfig struct {
	Enabled bool
}

type ColdConfig struct {
	Filecoin FilecoinConfig
}

type FilecoinConfig struct {
	Enabled    bool
	WalletAddr string
}

type CidConfig struct {
	ID         ConfigID
	InstanceID InstanceID
	Cid        cid.Cid
	Hot        HotConfig
	Cold       ColdConfig
}

type MinerSelector interface {
	GetTopMiners(n int) ([]MinerProposal, error)
}

type MinerProposal struct {
	Addr       string
	EpochPrice uint64
}

var (
	EmptyID = InstanceID("")
)

type InstanceID string

func (i InstanceID) Valid() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}
func (i InstanceID) String() string {
	return string(i)
}

func NewID() InstanceID {
	return InstanceID(uuid.New().String())
}

var (
	EmptyConfigID = ConfigID("")
)

type ConfigID string

func (i ConfigID) Valid() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}
func (i ConfigID) String() string {
	return string(i)
}

func NewConfigID() ConfigID {
	return ConfigID(uuid.New().String())
}

type CidInfo struct {
	ConfigID ConfigID
	Cid      cid.Cid
	Created  time.Time
	Hot      HotInfo
	Cold     ColdInfo
}

type HotInfo struct {
	Size int
	Ipfs IpfsHotInfo
}

type IpfsHotInfo struct {
	Created time.Time
}

type ColdInfo struct {
	Filecoin FilInfo
}

type FilInfo struct {
	PayloadCID cid.Cid
	Duration   uint64
	Proposals  []FilStorage
}

type FilStorage struct {
	ProposalCid cid.Cid
	Failed      bool
}
