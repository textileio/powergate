package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/go-units"
	"github.com/textileio/powergate/v2/ffs"
	"golang.org/x/sys/unix"
)

const (
	DiskSpaceCheck = "datacap"
)

type DiskSpaceAlert struct {
	JobID ffs.JobID
}

type diskSpaceAlertNotification struct {
	JobID              ffs.JobID `json:"jobId"`
	AlertType          string    `json:"alertType"`
	AvailableDiskSpace string    `json:"availableDiskSpace"`
	Error              string    `json:"error"`
}

func (d DiskSpaceAlert) Payload() (io.Reader, error) {
	availableDiskSpace, err := getAvailableDiskSpace()
	if err != nil {
		err = fmt.Errorf("failed to get available disk space: %w", err)
		log.Error(err)
		return nil, err
	}

	obj := &diskSpaceAlertNotification{
		JobID:              d.JobID,
		AlertType:          DiskSpaceCheck,
		AvailableDiskSpace: units.BytesSize(float64(availableDiskSpace)),
		Error:              "available disk space below threshold",
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}

func (d DiskSpaceAlert) MatchEvent(event string) bool {
	return false
}

func (d DiskSpaceAlert) MatchAlert(alert *ffs.WebhookAlert) bool {
	if alert == nil {
		return false
	}

	if alert.Type != DiskSpaceCheck {
		return false
	}

	threshold, err := parseDiskSpaceThresholdToBytes(alert.Threshold)
	if err != nil {
		log.Errorf("failed to parse alert disk space threshold: %s", err)
		return false
	}

	availableDiskSpace, err := getAvailableDiskSpace()
	if err != nil {
		log.Errorf("failed to get available disk space: %s", err)
		return false
	}

	return availableDiskSpace < threshold
}

func parseDiskSpaceThresholdToBytes(threshold string) (uint64, error) {
	size, err := units.RAMInBytes(threshold)
	if err != nil {
		return 0, err
	}

	return uint64(size), nil
}

// getAvailableDiskSpace - provides available disk space in bytes
func getAvailableDiskSpace() (uint64, error) {
	var stat unix.Statfs_t

	wd, err := os.Getwd()
	if err != nil {
		return 0, err
	}

	if err := unix.Statfs(wd, &stat); err != nil {
		return 0, err
	}

	// available blocks * size per block
	return stat.Bavail * uint64(stat.Bsize), nil
}
