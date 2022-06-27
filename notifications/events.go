package notifications

const (
	all           = "*"
	created       = "created"
	completed     = "completed"
	retried       = "retried"
	failed        = "failed"
	canceled      = "canceled"
	expired       = "expired"
	slashed       = "slashed"
	separator     = "-"
	storageDeal   = "storage-deal"
	dataRetrieval = "data-retrieval"

	// All events
	AllEvents          = all
	AllCreatedEvents   = all + separator + created
	AllCompletedEvents = all + separator + completed
	AllRetriedEvents   = all + separator + retried
	AllFailedEvents    = all + separator + failed
	AllCanceledEvents  = all + separator + canceled

	// Storage deal events
	AllStorageDealEvents      = storageDeal + separator + all
	StorageDealCreatedEvent   = storageDeal + separator + created
	StorageDealCompletedEvent = storageDeal + separator + completed
	StorageDealRetriedEvent   = storageDeal + separator + retried
	StorageDealFailedEvent    = storageDeal + separator + failed
	StorageDealCanceledEvent  = storageDeal + separator + canceled
	StorageDealExpiredEvent   = storageDeal + separator + expired
	StorageDealSlashedEvent   = storageDeal + separator + slashed

	// Data retrieval events
	AllDataRetrievalEvents      = dataRetrieval + separator + all
	DataRetrievalCreatedEvent   = dataRetrieval + separator + created
	DataRetrievalCompletedEvent = dataRetrieval + separator + completed
	DataRetrievalRetriedEvent   = dataRetrieval + separator + retried
	DataRetrievalFailedEvent    = dataRetrieval + separator + failed
	DataRetrievalCanceledEvent  = dataRetrieval + separator + canceled
)
