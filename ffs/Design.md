# FFS design

## Overview

This document presents the general design of the `fps` package of `fil-tools`.

**Disclaimer**: This's ongoing work, so some design, interface definition, etc. might change soon as implementation continues. 

The following picture presents principal packages and interfaces that are part of the design:
![FFS Design](https://user-images.githubusercontent.com/6136245/76028631-c7d61400-5f11-11ea-8234-c0cd143a142b.png)


The picture has an advanced scenario where different _Powergate_ instances are wired to different _Scheduler_ instances. Component names prefixed with * don't exist but are mentioned as possible implementations of existing interfaces.

A _Scheduler_ instance receives tasks from _Powergate_ instances to store a _Cid_ with a particular _CidConfig_ (e.g: only in hot layer, only in cold layer, in hot and cold layer only in 1 miner, etc). 

Each task is transformed into a _Job_ since it's execution should be async. This _Job_ state changes when the _Scheduler_ delegates work to the _Hot Layer_ and _Cold Layer_ configuration. The _Powergate_ cant _watch_ for changes in existing _Jobs_ to keep track of what's happening with a particular ongoing Cid configuration that was pushed.

## Components
The following sections give a more detailed description of each component and interface in the diagram.

### Manager
This component is responsible for creating _Powergate_ instances. On every creation, an _auth-token_ is created, which is mapped to that instance. The client uses this _auth-token_ in each request in the API so that the _Manager_ can redirect the action to its corresponding _Powergate_ instance. The _auth-token_ to _Powergate_ mapping is controlled by the _Auth_ component. Any further improvement regarding _auth-token_ permissions, invalidations, creating multiple ones, etc. should evolve in the _Auth_ component.

The _Manager_ uses a _WalletManager_ to generate a new account for a newly created _Powergate_ instance. The current _WalletModule_ implementation of _WalletManager_ supports using a _master-addr_ to automatically send configured funds to every newly created wallet address, mostly interested in making _Powergate_ have enough funds to be usable.

### Powergate
_Powergate_ is a concrete instance of the FFS. It deals with managing the following data:
- _Wallet Addr_ owned by the instance, and it's balance information.
- _CidConfigurations_, which represents the desired state for a particular Cid to be stored.
- _CidInformation_ which represents the current storing state for a particular stored Cid.

Each action of storing a new _Cid_ with a particular desired _CidConfiguration_ is pushed to the _Scheduler_, and a _JobID_ is returned to keep track of that async work. The _Powergate_ subscribes to the _Scheduler_ to watch for state changes of that _Job_.

### Scheduler

A _Scheduler_ is a component that _Powergate_ instances pushes new _CidConfigurations_ enforced for a _Cid_. These _CidConfigurations_ contain information about how to store that _Cid_ in the _Hot Layer_ and _Cold Layer_.

The _Scheduler_ doesn't know about the particular implementations of _Hot Layer_ or _Cold Layer_, only relies on interfaces:

```go
// HotLyer is a fast datastorage layer for storing and retrieving raw
// data or Cids.
type HotLayer interface {
    Add(context.Context, io.Reader) (cid.Cid, error)
    Get(context.Context, cid.Cid) (io.Reader, error)
    Pin(context.Context, cid.Cid) (HotInfo, error)
}

// ColdLayer is a slow datastorage layer for storing Cids.
type ColdLayer interface {
    Store(ctx context.Context, c cid.Cid, conf ColdConfig) (ColdInfo, error)
}
```

It also relies on a _MinerSelector_ interfaces which implement a particular strategy to fetch the most desirable N miners needed for making deals in the _Cold Layer_:
```go
// MinerSelector returns miner addresses and ask storage information using a
// desired strategy.
type MinerSelector interface {
    GetTopMiners(n int) ([]MinerProposal, error)
}
```
Particular implementations of _MinerSelector_ includes:
- _FixedMiners_: which always returns a particular fixed list of miner addresses.
- _ReputationSorted_: which returns the miner addresses using a reputation system built on top of miner information.

In summary, a _Scheduler_ instance act differently depending on which instances on its _Hot Layer_, _Cold Layer_, and _Miner Selector_ implementations. In the diagram above shows two configurations (surrounded by dotted boxes).

In the first dotted box, a _Scheduler_ uses an _IPFS Node_ as the _HotLayer_ using the _CoreIPFS_ adapter as the interface implementation, which uses the _http api_ client to talk with the _IPFS node_. It also uses the _ColdFil_ adapter as the _ColdLayer_ implementation, which uses the _DealModule_ to make deals with a _Lotus instance_. It uses a _ReputationSorted_ implementation of _MinerSelector_ to fetch the best miners from a miner's reputation system.

In the second dotted box, shows another possible configuration in which uses an _IPFS Cluster_ with a _HotIpfsCluster_ adapter of _HotLayer_; or even a more advanced _HotLayer_ called _HotS3IpfsCluster_ which saves _Cid_ into _IPFS Cluster_ and some _AWS S3_ instance. The _MinerSelector_ implementation for the _ColdLayer_ is _FixedMiners_ which always returns a configured fixed list of miners to make deals with.

Finally, _Powergate_ instances are wired to different _Scheduler_ instances depending on which _Scheduler_-configuration may suit better for them. This is a possibility of the current design.

### Powergate <-> Scheduler
Considering the _Scheduler_ interface:
```go
// Scheduler creates and manages Job which executes Cid configurations
// in Hot and Cold layers, enables retrieval from those layers, and
// allows watching for Job state changes.
// (TODO: Still incomplete for retrieval apis and rough edges)
type Scheduler interface {
    Enqueue(CidConfig) (JobID, error)
    GetFromHot(ctx context.Context, c cid.Cid) (io.Reader, error)
    GetJob(JobID) (Job, error)

    Watch(InstanceID) <-chan Job
    Unwatch(<-chan Job)
}
```

When a _Powergate_ instance receives the order to pin a _Cid_ with a particular configuration, it _Enqueues_ the desired configuration on the _Scheduler_, immediately getting back a _JobID_ for that action.

The _Powergate_ instance can pull the current state of the created _Job_, or it can _Watch_ all _Job_ changes corresponding to that instance to avoid polling. _Job_ status (_JobStatus_) include: _Queued_, _InProgress_, _Failed_, _Cancelled_, _Done_. In the case of _JobStatus_ is _Failed_, in _Job.ErrCause_ is the cause of the error.