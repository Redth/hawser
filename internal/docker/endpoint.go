package docker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
)

const (
	DefaultUnixSocket  = "/var/run/docker.sock"
	DefaultNpipeSocket = `\\.\pipe\docker_engine`
)

// Endpoint describes how Hawser connects to the Docker Engine API.
type Endpoint struct {
	Scheme  string
	Address string
	Raw     string
	BaseURL string
}

// ParseEndpoint normalizes Docker socket/host values such as Unix sockets,
// Windows named pipes, and TCP Docker hosts.
func ParseEndpoint(value string) Endpoint {
	value = strings.TrimSpace(value)
	if value == "" {
		if runtime.GOOS == "windows" {
			value = DefaultNpipeSocket
		} else {
			value = DefaultUnixSocket
		}
	}

	lower := strings.ToLower(value)
	switch {
	case strings.HasPrefix(lower, "unix://"):
		address := strings.TrimPrefix(value, value[:len("unix://")])
		return Endpoint{Scheme: "unix", Address: address, Raw: "unix://" + address, BaseURL: "http://docker"}
	case strings.HasPrefix(lower, "npipe://"):
		address := npipeURLToPath(value)
		return Endpoint{Scheme: "npipe", Address: address, Raw: npipePathToURL(address), BaseURL: "http://docker"}
	case strings.HasPrefix(lower, "tcp://"):
		address := strings.TrimPrefix(value, value[:len("tcp://")])
		return Endpoint{Scheme: "tcp", Address: address, Raw: "tcp://" + address, BaseURL: "http://" + address}
	case strings.HasPrefix(lower, "http://"):
		address := strings.TrimPrefix(value, value[:len("http://")])
		return Endpoint{Scheme: "http", Address: address, Raw: value, BaseURL: value}
	case strings.HasPrefix(lower, "https://"):
		address := strings.TrimPrefix(value, value[:len("https://")])
		return Endpoint{Scheme: "https", Address: address, Raw: value, BaseURL: value}
	case isNpipePath(value):
		return Endpoint{Scheme: "npipe", Address: value, Raw: npipePathToURL(value), BaseURL: "http://docker"}
	default:
		return Endpoint{Scheme: "unix", Address: value, Raw: "unix://" + value, BaseURL: "http://docker"}
	}
}

// DockerHostEnv returns a DOCKER_HOST environment variable for Docker CLI calls.
func DockerHostEnv(value string) string {
	endpoint := ParseEndpoint(value)
	return "DOCKER_HOST=" + endpoint.Raw
}

func (e Endpoint) DialContext(ctx context.Context) (net.Conn, error) {
	var dialer net.Dialer
	switch e.Scheme {
	case "unix":
		return dialer.DialContext(ctx, "unix", e.Address)
	case "npipe":
		return dialNpipe(ctx, e.Address)
	case "tcp", "http", "https":
		return dialer.DialContext(ctx, "tcp", e.Address)
	default:
		return nil, fmt.Errorf("unsupported Docker endpoint scheme: %s", e.Scheme)
	}
}

func (e Endpoint) Transport() *http.Transport {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     0,
	}

	if e.Scheme == "https" {
		return transport
	}

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return e.DialContext(ctx)
	}
	return transport
}

func isNpipePath(value string) bool {
	return strings.HasPrefix(value, `\\.\pipe\`) || strings.HasPrefix(value, `\\?\pipe\`)
}

func npipeURLToPath(value string) string {
	path := strings.TrimPrefix(value, value[:len("npipe://")])
	path = strings.TrimLeft(path, "/")
	path = strings.ReplaceAll(path, "/", `\`)
	if strings.HasPrefix(path, `\\`) {
		return path
	}
	return `\\` + path
}

func npipePathToURL(path string) string {
	trimmed := strings.TrimPrefix(path, `\\`)
	trimmed = strings.ReplaceAll(trimmed, `\`, "/")
	return "npipe:////" + trimmed
}
