# FFS design

## Overview

This document presents the general design of the `ffs` package of `powergate`.

**Disclaimer**: This's ongoing work, so some design, interface definition, etc. might change soon as implementation continues. 

The following picture presents principal packages and interfaces that are part of the design:
![FFS Design](https://user-images.githubusercontent.com/6136245/79258737-fbd21c80-7e61-11ea-9616-3521f543184f.png)


The picture has an advanced scenario where different _API_ instances are wired to different _Scheduler_ instances. Component names prefixed with * don't exist but are mentioned as possible implementations of existing interfaces.

The central idea about the design is that an _API_ defines the desired storing state for a _Cid_ using a _CidConfig_ struct. This struct has information about desired storing state configuration in the Hot and Cold storages.

When a new or updated _CidConfig_ is pushed in an _API_, it delegates this work to the _Scheduler_. The _Scheduler_ will execute whatever work is necessary to comply with the new/updated Cid configuration.

From the _Scheduler_ point of view, this work is considered a _Job_ created by _API_. The job refers to doing the necessary work to enforce the new _CidConfig_. The _API_ can watch for this _Job_ state changes to see if the task of pushing a new _CidConfig_ is queued, in progress, executing, executed successfully, or failed. The _Scheduler_ also provides a human-friendly log stream of work being done for a _Cid_.

Apart from executing _API_ triggered work, like pushing a new _CidConfig_, the _Scheduler_ also does some background jobs related to deals-renewal for Cids, which have this feature enabled in their _CidConfig_. Similarly, it has background jobs for repair actions.

## Components
The following sections give a more detailed description of each component and interface in the diagram.

### Manager
This component is responsible for creating _API_ instances. When a new _API_ instance is created, an _auth-token_ for this instance is also created. The client uses this _auth-token_ in each request in the API so that the _Manager_ can redirect the action to its corresponding _API_ instance, while also having some minimal access-control validation.

The mapping between _auth-tokens_ and _API_ is controlled by an _Auth_ component. Further features such as token invalidation, finer-grained access control per action, or multiple auth token support will live in this module.

Every _API_ instance needs a dedicated Filecoin address that will be used to pay for actions done on the network. _Manager_ delegates wallet related activities to _WalletManager_, such as: creating new addresses for new _API_ instances, sending funds to those addresses, getting the balance.

### API
_API_ is a concrete instance of FFS to be used by a client.
It owns the following information:
- A Filecoin address.
- _CidConfigs_ describing the desired state for Cids to be stored in Hot and Cold storage.
- A default _CidConfig_ to be used unless an explicit _CidConfig_ is given.

It has APIs to create/update _CidConfigs_, get its address information such as balance, watch for _Job_ state changes or human-friendly Log outputs about work done by the _Scheduler_. Refer to the _CidConfig_ section to understand more about this important structure.

### Scheduler

The main goal of this component is to do whatever its possible to reach a desired storing state for a _Cid_.

Its interface for _API_ is defined by the interface:
```go
// Scheduler enforces a CidConfig orchestrating Hot and Cold storages.
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

    // WatchJobs sends to a channel state updates for all Jobs created by
    // an Instance.
    WatchJobs(context.Context, chan<-Job, APIID) error

    // WatchLogs writes new log entries from Cid related executions.
    // This is a blocking operation that should be canceled by canceling the
    // provided context.
    WatchLogs(context.Context, chan<- LogEntry) error
}
```

### Responsibilities
When a new/updated _CidConfig_ is pushed by an _API_, the _Scheduler_ bounds the work of enforcing that state in a _Job_.
This _Job_ has a lifecycle: queued, in progress, done, or failed.

Apart from _Jobs_, the _Scheduler_ has background tasks that monitor deal renewals or repair operations.

In summary, the _Scheduler_ is concerned about enforcing a _CidConfig_ for a Cid. It does this by inspecting the current state of the Cid in both storages, deciding on which is the necessary actions to make in both layers, and using the Hot and Cold storage APIs to execute that necessary work. 

#### Hot and Cold storage abstraction
The _Scheduler_ is abstracted from particular implementations of the _Hot Storage_ and _Cold Storage_.
It relies on the following interfaces:

```go
// HotStorage is a fast datastorage layer for storing and retrieving raw
// data or Cids.
type HotStorage interface {
    Add(context.Context, io.Reader) (cid.Cid, error)
    Remove(context.Context, cid.Cid) error
    Get(context.Context, cid.Cid) (io.Reader, error)
    Store(context.Context, cid.Cid) (int, error)
    Put(context.Context, blocks.Block) error
    IsStored(context.Context, cid.Cid) (bool, error)
}

// ColdStorage is a slow datastorage layer for storing Cids.
type ColdStorage interface {
    Store(context.Context, cid.Cid, string, FilConfig) (FilInfo, error)
    Retrieve(context.Context, cid.Cid, car.Store, string) (cid.Cid, error)

    EnsureRenewals(context.Context, cid.Cid, FilInfo, string, FilConfig) (FilInfo, error)
    IsFilDealActive(context.Context, cid.Cid) (bool, error)
}
```

#### MinerSelector abstraction
It also relies on a _MinerSelector_ interfaces which implement a particular strategy to fetch the most desirable N miners needed for making deals in the _Cold Storage_:
```go
// MinerSelector returns miner addresses and asks storage information using a
// desired strategy.
type MinerSelector interface {
    GetMiners(int, MinerSelectorFilter) ([]MinerProposal, error)
}
```
Particular implementations of _MinerSelector_ include:
- _FixedMiners_: which always returns a particular fixed list of miner addresses.
- _ReputationSorted_: which returns the miner addresses using a reputation system built on top of miner information.


#### Configuration scenarios
Looking at diagram in the _Overview_ section  we can see some different Hot and Cold storages:

In the first dotted box, a _Scheduler_ uses an _IPFS Node_ as the _HotStorage_ using the _CoreIPFS_ adapter as the interface implementation, which uses the _http api_ client to talk with the _IPFS node_. It also uses the _ColdFil_ adapter as the _ColdStorage_ implementation, which uses the _DealModule_ to make deals with a _Lotus instance_. It uses a _ReputationSorted_ implementation of _MinerSelector_ to fetch the best miners from a miner's reputation system.

In the second dotted box, shows another possible configuration in which uses an _IPFS Cluster_ with a _HotIpfsCluster_ adapter of _HotStorage_; or even a more advanced _HotStorage_ called _HotS3IpfsCluster_ which saves _Cid_ into _IPFS Cluster_ and some _AWS S3_ instance. The _MinerSelector_ implementation for the _ColdStorage_ is _FixedMiners_ which always returns a configured fixed list of miners to make deals with.



### Cid Configuration
Cid configurations are a central part of FFS mechanics. An _API_ defines the desired state of the Cid in the Hot and Cold storage. Currently, it has the following structure:
```go
// CidConfig has a Cid desired storing configuration for a Cid in
// Hot and Cold storage.
type CidConfig struct {
    // Cid is the Cid of the stored data.
    Cid cid.Cid
    // Hot has desired storing configuration in Hot Storage.
    Hot HotConfig
    // Cold has desired storing configuration in the Cold Storage.
    Cold ColdConfig
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

// IpfsConfig is the desired storage of a Cid in IPFS.
type IpfsConfig struct {
    // AddTimeout is an upper bound on adding data to IPFS node from
    // the network before failing.
    AddTimeout int
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
}

// FilRenew contains renew configuration for a Cid Cold Storage deals.
type FilRenew struct {
    // Enabled indicates that deal-renewal is enabled for this Cid.
    Enabled bool
    // Threshold indicates how many epochs before expiring should trigger
    // deal renewal. e.g: 100 epoch before expiring.
    Threshold int
}
```

Each attribute has a description of its goal.

Both the Hot and Cold configurations have an `Enable` flag to enable/disable the Cid data storage in each of them.
If a client only wants to save data in the Cold storage, it can set `HotConfig.Enabled: false` and `ColdConfig.Enabled: true`. The same applies inversely.

#### _API_ _Get(...)_ operation
One important point is that `Get` operations in _API_ can only retrieve data from the Hot Storage (via `GetCidFromHot` in the _Scheduler_).
This has some different scenarios:
- If the data is stored in the Hot Storage, it fetched from there.
- If the data wasn't enabled in the Hot Storage (`HotConfig.Enabled: false`), it will error indicating that the Hot Storage isn't enabled.

The last point indicates that the _API_ client should explicitly set `HotConfig.Enabled: true` to be able to retrieve the data. Hot Storage enabling is done in two steps:
1) It tries to fetch the data from the IPFS network considering the `AddTimeout` as a bound of time.
2) If the last step failed:
2.a) If `HotConfig.AllowUnfreeze: false`, it fails since it couldn't fetch the data from the single allowed source (IPFS network).
2.b) If `HotConfig.AllowUnfreeze: true`; it will check if the data is available at Cold Storage. If that's the case, it will _unfreeze the data_, and save it to Hot Storage. This allows a `Get` operation afterward.

The rationale behind asking the client to enable hot storage with allow-unfreeze is related to the fact that retrieving data from Filecion incurs in an economic cost that will be paid by the _API_ address. Retrieving data from the IPFS network is considered _free_ (discarding unavoidable bandwidth costs, etc).

### Updating CidConfig
The _Scheduler_ is always checking the current state of Cid storage before executing actions regarding an updated _CidConfig_.

_CidConfig_ changes regarding Hot Storage are always applied since Hot Storage is, in general, malleable. In particular, enabling or disabling Hot Storage is most probably easy to execute and thus have a predictable ending state.

_CidConfig_ changes regarding Cold Storage have more subtle meaning. For example, if the _RepFactor_ is increased the _Scheduler_ will be aware of the current _RepFactor_ and only make enough new deals to ensure its new value. e.g: if _RepFactor_ was 1 and the updated _CidConfig_ has _RepFactor_ 2, it will only make one new deal. 
As another example, if _RepFactor_ was 2 and is decreased to 1, the _Scheduler_ won't execute any actual work since one of the two current active deals will eventually expire.

The _RepFactor_ configuration also is considered if the Cid has enabled automatic deal renweal. In particular, if the _RepFactor_ was decreased from 3 to 1, the rewneal logic will wait until the last deal is close to expiring to only renew that one. That's saying, the renew logic doesn't blindly renew expiring deals, but it's _RepFactor aware_ as expected.

Regarding other Cold Storage configuration changes regarding miner selection, such as country filtering or excluded miners, these new considerations will be made every time a new deal is made. Any other existing deals that are active that were created on other configuration conditions can't be canceled or reverted. Saying it differently, the new miner-related configuration will be considered from future new deals, i.e: renewing deals, increased _RepFactor_, repairing.
