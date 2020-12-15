package cmd

import (
	"fmt"
	"sync"
)

// StagingStatus tracks the available resources on staging.
type StagingStatus struct {
	cachedBytes     int64
	maxStagingBytes int64
	minDealBytes    int64
	m               sync.Mutex
}

// Ready returns true if there is space on the remote staging endpoint.
func (sc *StagingStatus) Ready(bytes int64) (bool, error) {
	sc.m.Lock()
	defer sc.m.Unlock()
	if bytes > sc.maxStagingBytes {
		err := fmt.Errorf("request larger than available staging limit: %d needed out of %d", bytes, sc.maxStagingBytes)
		return false, err
	}
	if sc.cachedBytes+bytes > sc.maxStagingBytes {
		return false, nil
	}
	sc.cachedBytes += bytes
	return true, nil
}

// Done frees space being held on staging for a task.
func (sc *StagingStatus) Done(bytes int64) {
	sc.m.Lock()
	defer sc.m.Unlock()
	sc.cachedBytes -= bytes
}
