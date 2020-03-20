# FFS design

## Overview

This document presents the general design of the `fps` package of `powergate`.

**Disclaimer**: This's ongoing work, so some design, interface definition, etc. might change soon as implementation continues. 

The following picture presents principal packages and interfaces that are part of the design:
![FFS Design](https://user-images.githubusercontent.com/6136245/76992075-6ce8e780-6929-11ea-9f23-f90f1c6bffe7.png)


The picture has an advanced scenario where different _Api_ instances are wired to different _Scheduler_ instances. Component names prefixed with * don't exist but are mentioned as possible implementations of existing interfaces.

A _Scheduler_ instance receives tasks from _Api_ instances to store a _Cid_ with a particular _CidConfig_ (e.g: only in hot layer, only in cold layer, in hot and cold layer only in 1 miner, etc). 

Each task is transformed into a _Job_ since it's execution should be async. This _Job_ state changes when the _Scheduler_ delegates work to the _Hot Storage_ and _Cold Storage_ configuration. The _Api_ cant _watch_ for changes in existing _Jobs_ to keep track of what's happening with a particular ongoing Cid configuration that was pushed.

## Components
The following sections give a more detailed description of each component and interface in the diagram.

### Manager
This component is responsible for creating _Api_ instances. On every creation, an _auth-token_ is created, which is mapped to that instance. The client uses this _auth-token_ in each request in the API so that the _Manager_ can redirect the action to its corresponding _Api_ instance. The _auth-token_ to _Api_ mapping is controlled by the _Auth_ component. Any further improvement regarding _auth-token_ permissions, invalidations, creating multiple ones, etc. should evolve in the _Auth_ component.

The _Manager_ uses a _WalletManager_ to generate a new account for a newly created _Api_ instance. The current _WalletModule_ implementation of _WalletManager_ supports using a _master-addr_ to automatically send configured funds to every newly created wallet address, mostly interested in making _Api_ have enough funds to be usable.

### Api
_Api_ is a concrete instance of the FFS. It deals with managing the following data:
- _Wallet Addr_ owned by the instance, and it's balance information.
- _CidConfigurations_, which represents the desired state for a particular Cid to be stored.
- _CidInformation_ which represents the current storing state for a particular stored Cid.

Each action of storing a new _Cid_ with a particular desired _CidConfiguration_ is pushed to the _Scheduler_, and a _JobID_ is returned to keep track of that async work. The _Api_ subscribes to the _Scheduler_ to watch for state changes of that _Job_.

### Scheduler

A _Scheduler_ is a component that _Api_ instances pushes new _CidConfigurations_ enforced for a _Cid_. These _CidConfigurations_ contain information about how to store that _Cid_ in the _Hot Storage_ and _Cold Storage_.

The _Scheduler_ doesn't know about the particular implementations of _Hot Storage_ or _Cold Storage_, only relies on interfaces:

```go
// HotStorage is a fast datastorage layer for storing and retrieving raw
// data or Cids.
type HotStorage interface {
	Add(context.Context, io.Reader) (cid.Cid, error)
	Get(context.Context, cid.Cid) (io.Reader, error)
	Pin(context.Context, cid.Cid) (int, error)
	Put(context.Context, blocks.Block) error
}

// ColdStorage is a slow datastorage layer for storing Cids.
type ColdStorage interface {
	Store(context.Context, cid.Cid, string, FilConfig) (FilInfo, error)
	Retrieve(context.Context, cid.Cid, car.Store, string) (cid.Cid, error)

	EnsureRenewals(context.Context, cid.Cid, FilInfo, string, FilConfig) (FilInfo, error)
}
```

It also relies on a _MinerSelector_ interfaces which implement a particular strategy to fetch the most desirable N miners needed for making deals in the _Cold Storage_:
```go
// MinerSelector returns miner addresses and ask storage information using a
// desired strategy.
type MinerSelector interface {
	GetMiners(int, MinerSelectorFilter) ([]MinerProposal, error)
}
```
Particular implementations of _MinerSelector_ includes:
- _FixedMiners_: which always returns a particular fixed list of miner addresses.
- _ReputationSorted_: which returns the miner addresses using a reputation system built on top of miner information.

In summary, a _Scheduler_ instance act differently depending on which instances on its _Hot Storage_, _Cold Storage_, and _Miner Selector_ implementations. In the diagram above shows two configurations (surrounded by dotted boxes).

In the first dotted box, a _Scheduler_ uses an _IPFS Node_ as the _HotStorage_ using the _CoreIPFS_ adapter as the interface implementation, which uses the _http api_ client to talk with the _IPFS node_. It also uses the _ColdFil_ adapter as the _ColdStorage_ implementation, which uses the _DealModule_ to make deals with a _Lotus instance_. It uses a _ReputationSorted_ implementation of _MinerSelector_ to fetch the best miners from a miner's reputation system.

In the second dotted box, shows another possible configuration in which uses an _IPFS Cluster_ with a _HotIpfsCluster_ adapter of _HotStorage_; or even a more advanced _HotStorage_ called _HotS3IpfsCluster_ which saves _Cid_ into _IPFS Cluster_ and some _AWS S3_ instance. The _MinerSelector_ implementation for the _ColdStorage_ is _FixedMiners_ which always returns a configured fixed list of miners to make deals with.

Finally, _Api_ instances are wired to different _Scheduler_ instances depending on which _Scheduler_-configuration may suit better for them. This is a possibility of the current design.

### Api <-> Scheduler
Considering the _Scheduler_ interface:
```go
/ Scheduler enforces a CidConfig orchestrating Hot and Cold storages.
type Scheduler interface {
	// PushConfig push a new or modified configuration for a Cid. It returns
	// the JobID which tracks the current state of execution of that task.
	PushConfig(PushConfigAction) (JobID, error)

	// GetCidInfo returns the current Cid storing state. This state may be different
	// from CidConfig which is the *desired* state.
	GetCidInfo(cid.Cid) (CidInfo, error)
	// GetCidFromHot returns an Reader with the Cid data. If the data isn't in the Hot
	// Storage, it errors with ErrHotStorageDisabled.
	GetCidFromHot(context.Context, cid.Cid) (io.Reader, error)

	// GetJob gets the a Job.
	GetJob(JobID) (Job, error)

	// Watch returns a channel which will receive updates for all Jobs created by
	// an Instance.
	Watch(ApiID) <-chan Job
	// Unwatch unregisters a subscribed channel.
	Unwatch(<-chan Job)
}
```

When a _Api_ instance receives the order to pin a _Cid_ with a particular configuration, it _Enqueues_ the desired configuration on the _Scheduler_, immediately getting back a _JobID_ for that action.

The _Api_ instance can pull the current state of the created _Job_, or it can _Watch_ all _Job_ changes corresponding to that instance to avoid polling. _Job_ status (_JobStatus_) include: _Queued_, _InProgress_, _Failed_, _Cancelled_, _Done_. In the case of _JobStatus_ is _Failed_, in _Job.ErrCause_ is the cause of the error.
