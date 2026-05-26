//go:build windows

package metrics

// collectDisk is not implemented for Windows yet; Dockhand displays zero values as N/A.
func (c *Collector) collectDisk() (total, used, free uint64, err error) {
	return 0, 0, 0, nil
}
