# FFS design

## Overview

This document presents the general design of the `ffs` package of `powergate`.

**Disclaimer**: This's ongoing work so the design will continue to change.

The following picture presents principal packages and interfaces that are part of the design:
![FFS Design](https://user-images.githubusercontent.com/6136245/83649396-847d5700-a58d-11ea-8d93-5ea20ca1bda7.png)


The picture has an advanced scenario where different _API_ instances are wired to different _Scheduler_ instances. Component names prefixed with * don't exist but are mentioned as possible implementations of existing interfaces.

The central idea about the design is that an _API_ defines the desired storing state for a _Cid_ using a _CidConfig_ struct. This struct has information about desired storing state configuration in the Hot and Cold storages.

When a new or updated _CidConfig_ is pushed in an _API_, it delegates this work to the _Scheduler_. The _Scheduler_ will execute whatever work is necessary to comply with the new/updated Cid configuration.

From the _Scheduler_ point of view, this work is considered a _Job_ created by _API_. The job refers to doing the necessary work to enforce the new _CidConfig_. The _API_ can watch for this _Job_ state changes to see if the task of pushing a new _CidConfig_ is queued, executing, finished successfully, failed, or canceled. The _Scheduler_ also provides a human-friendly log stream of work being done for a _Cid_.

The _Scheduler_ also executes proactive actions for prior pushed _CidConfigs_ which enabled the _renew_ or _repair_ feature. Finally, the _Scheduler_ is designed to resume any kind of interrupted job executions.

## Components
The following sections give a more detailed description of each component and interface in the diagram.

### Manager
This component is responsible for creating _API_ instances. When a new _API_ instance is created, an _auth-token_ for this instance is also created. The client uses this _auth-token_ in each request in the API so that the _Manager_ can redirect the action to its corresponding _API_ instance, while also having some minimal access-control validation.

The mapping between _auth-tokens_ and _API_ is controlled by an _Auth_ component. Further features such as token invalidation, finer-grained access control per action, or multiple auth token support will live in this module.

Since _API_ might store data in the Filecoin network, they're asigned a newly created Filecoin address which will be controlled by the underlying Filecoin client used in the _ColdStorage_. The process of creating and assigning this new wallet account is done automatically by _Manager_, using a subcomponent _WalletManager_.

_Manager_ enables being configured to auto-fund newly created wallet addresses, so new created _API_ can have funds to execute actions in the Filecoin network. This feature can be optionally enabled. If enabled, a _masterAddress_ and _initialFunds_ will be configured which indicates from which Filecoin Client wallet address funds will be sent and the amount of the transfer.


### API
_API_ is a concrete instance of FFS to be used by a client.
It owns the following information:
- At least one Filecoin address. Later the client can opt to create more address and indicate which to use when making action.
- _CidConfigs_ describing the desired state for Cids to be stored in Hot and Cold storage.
- A default _CidConfig_ to be used unless an explicit _CidConfig_ is given.

The instance provides apis to:
- Get and Set the default _CidConfig_ used to store new data.
- Get summary information about all the _Cid_ stored in this instance.
- Manage Filecoin wallet addresses under its control.
- Sending FIL transactions from owned Filecoin wallet addresses.
- Create, replace and remove _CidConfig_ which indicates which cids to store in the instance.
- Provide detailed information about a particular stored Cid.
- Get information about status of executing _Jobs_ corresponding to the FFS instance.
- Human-friendly log streams about events happening for a _Cid_, from storage, renewals, repair and anything related to actions being done for it.

### Scheduler

In a nutshell, the _Scheduler_ is the component responsible for orchestrating the Hot and Cold storage to enforce indicated _CidConfigs_ by connected _API_.

Refer to the [Go docs](https://pkg.go.dev/github.com/textileio/powergate/ffs/scheduler?tab=doc) to see its exported API.

### Responsibilities
When a new _CidConfig_ is pushed by an _API_, the _Scheduler_ is responsible for orchestrating whatever actions are necessary to enforce it with the Hot and Col storage.

Every new _CidConfig_, being the first or newer version for a Cid, is encapsulated in a _Job_. A _Job_ is the unit of work which the _Scheduler_ executes. _Jobs_ have different status: _Queued_, _Executing_, _Done_, _Failed_, and _Canceled_.

Apart from executing _Jobs_, the _Scheduler_ has background processes to keep enforcing configuration features that requires tracking. For example, if a _CidConfig_ has renewal or repair enabled, the _Scheduler_ is responsible for do necessary work as expected.
Apart from _Jobs_, the _Scheduler_ has background tasks that monitor deal renewals or repair operations.

In summary, _APIs_ delegates *the desired state for a Cid* and the _Scheduler_ is responsible for *ensuring that state is true* by orchestrating the Hot and Cold storage.

#### Hot and Cold storage abstraction
The _Scheduler_ interacts with abstractions for the Hot and Cold storage.
Refer to the Go docs of the [HotStorage](https://pkg.go.dev/github.com/textileio/powergate@v0.0.1-beta.6/ffs?tab=doc#HotStorage) and [ColdStorage](https://pkg.go.dev/github.com/textileio/powergate@v0.0.1-beta.6/ffs?tab=doc#ColdStorage) to understand their APIs.

It can be noticed that the _ColdStorage_ interface is quite biased towards using a _Filecoin client_ in the implementation, but this enables to include also other tiered cold storages if wanted if deal creation or retrieval may be wanted. Refer to the diagram at the top of this document to understand possible configurations.

The _ColdStorage_ relies on a _MinerSelector_ interface to query the universe of available miners to make new deals. Refer to the [Go doc](https://pkg.go.dev/github.com/textileio/powergate/ffs@v0.0.1-beta.6?tab=doc#MinerSelector) to understand its API.

Powergate has the _Reputation Module_ which leverages built indexes about miners data to provide a universe of available miners soreted by a chosen criteria. In a full run of FFS, the _ColdStorage_ is connected to a _MinerSelector_ with the _Reputation Module_ implementation. However, for integration tests a _FixedMiners_ miner selector is used to bound the universe of available miners for deals to desired values.

The _MinerSelector_ API already provides enough filtering configuration to force using or excluding particular miners. In general, other implementations than the default one should be used if the universe of available miners wants to be completely controled by design, and not by available miners on the connected Filecoin network.

### Cid Configuration
In the current document we've refered to _CidConfigs_ as a central concept in the FFS module. A _CidConfig_ indicates the desired storing state of a _Cid_ scoped in a _API_. Refer to the [Go docs](https://pkg.go.dev/github.com/textileio/powergate/ffs@v0.0.1-beta.6?tab=doc#CidConfig) to understand its rich configuration.


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
