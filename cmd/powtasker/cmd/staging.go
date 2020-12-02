package cmd

import (
	"fmt"
	"sync"
)

type StagingConfig struct {
	cachedBytes     int64
	maxStagingBytes int64
	minDealBytes    int64
	m               sync.Mutex
}

func (sc *StagingConfig) Ready(bytes int64) (bool, error) {
	sc.m.Lock()
	defer sc.m.Unlock()
	if bytes > sc.maxStagingBytes {
		err := fmt.Errorf("Request larger than available staging limit: %d needed out of %d.", bytes, sc.maxStagingBytes)
		return false, err
	}
	if sc.cachedBytes+bytes >= sc.maxStagingBytes {
		return false, nil
	}
	sc.cachedBytes += bytes
	return true, nil
}

func (sc *StagingConfig) Done(bytes int64) {
	sc.m.Lock()
	defer sc.m.Unlock()
	sc.cachedBytes -= bytes
}
