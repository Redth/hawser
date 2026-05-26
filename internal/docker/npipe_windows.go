//go:build windows

package docker

import (
	"context"
	"net"

	"github.com/Microsoft/go-winio"
)

func dialNpipe(ctx context.Context, path string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, path)
}
