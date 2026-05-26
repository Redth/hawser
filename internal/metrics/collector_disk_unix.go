//go:build !windows

package metrics

import (
	"context"
	"syscall"
)

// collectDisk reads disk usage for Docker data root.
func (c *Collector) collectDisk() (total, used, free uint64, err error) {
	dataRoot, err := c.dockerClient.GetDataRoot(context.Background())
	if err != nil {
		dataRoot = "/var/lib/docker"
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(dataRoot, &stat); err != nil {
		return 0, 0, 0, err
	}

	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bavail * uint64(stat.Bsize)
	used = total - free

	return total, used, free, nil
}
