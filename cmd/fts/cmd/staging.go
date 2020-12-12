package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

type StagingLimit struct {
	cachedBytes     int64
	maxStagingBytes int64
	minDealBytes    int64
	m               sync.Mutex
}

func (sc *StagingLimit) Ready(bytes int64) (bool, error) {
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

func (sc *StagingLimit) Done(bytes int64) {
	sc.m.Lock()
	defer sc.m.Unlock()
	sc.cachedBytes -= bytes
}

func stageData(task Task, pow *client.Client, config PipelineConfig) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	ctx = context.WithValue(ctx, client.AuthKey, config.token)
	if task.IsDir {
		return pow.Data.StageFolder(ctx, config.ipfsrevproxy, task.Path)
	}
	f, err := os.Open(task.Path)
	if err != nil {
		return "", err
	}
	stageRes, err := pow.Data.Stage(ctx, f)
	if err != nil {
		return "", err
	}
	err = f.Close()
	return stageRes.Cid, err
}

func getReplicationFactor(task Task, config PipelineConfig) int {
	repFactor := 1
	if task.storageConfig != nil && task.storageConfig.Cold.Enabled {
		repFactor = int(task.storageConfig.Cold.Filecoin.ReplicationFactor)
	} else if config.storageConfig != nil && config.storageConfig.Cold.Enabled {
		repFactor = int(config.storageConfig.Cold.Filecoin.ReplicationFactor)
	}
	return repFactor
}

func applyConfig(task Task, pow *client.Client, config PipelineConfig) (*userPb.ApplyStorageConfigResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	ctx = context.WithValue(ctx, client.AuthKey, config.token)
	options := []client.ApplyOption{}
	if task.storageConfig != nil {
		options = append(options, client.WithStorageConfig(task.storageConfig))
	} else if config.storageConfig != nil {
		options = append(options, client.WithStorageConfig(config.storageConfig))
	}
	return pow.StorageConfig.Apply(ctx, task.CID, options...)
}
