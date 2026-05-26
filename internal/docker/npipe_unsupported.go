//go:build !windows

package docker

import (
	"context"
	"fmt"
	"net"
)

func dialNpipe(ctx context.Context, path string) (net.Conn, error) {
	return nil, fmt.Errorf("Windows named pipes are not supported on this platform: %s", path)
}
