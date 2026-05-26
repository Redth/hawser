package docker

import (
	"context"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		scheme  string
		address string
		raw     string
		baseURL string
	}{
		{
			name:    "unix socket path",
			input:   "/var/run/docker.sock",
			scheme:  "unix",
			address: "/var/run/docker.sock",
			raw:     "unix:///var/run/docker.sock",
			baseURL: "http://docker",
		},
		{
			name:    "unix endpoint",
			input:   "unix:///tmp/docker.sock",
			scheme:  "unix",
			address: "/tmp/docker.sock",
			raw:     "unix:///tmp/docker.sock",
			baseURL: "http://docker",
		},
		{
			name:    "tcp endpoint",
			input:   "tcp://127.0.0.1:2375",
			scheme:  "tcp",
			address: "127.0.0.1:2375",
			raw:     "tcp://127.0.0.1:2375",
			baseURL: "http://127.0.0.1:2375",
		},
		{
			name:    "npipe endpoint",
			input:   "npipe:////./pipe/docker_engine",
			scheme:  "npipe",
			address: `\\.\pipe\docker_engine`,
			raw:     "npipe:////./pipe/docker_engine",
			baseURL: "http://docker",
		},
		{
			name:    "npipe path",
			input:   `\\.\pipe\docker_engine`,
			scheme:  "npipe",
			address: `\\.\pipe\docker_engine`,
			raw:     "npipe:////./pipe/docker_engine",
			baseURL: "http://docker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := ParseEndpoint(tt.input)
			if endpoint.Scheme != tt.scheme {
				t.Fatalf("Scheme = %q, want %q", endpoint.Scheme, tt.scheme)
			}
			if endpoint.Address != tt.address {
				t.Fatalf("Address = %q, want %q", endpoint.Address, tt.address)
			}
			if endpoint.Raw != tt.raw {
				t.Fatalf("Raw = %q, want %q", endpoint.Raw, tt.raw)
			}
			if endpoint.BaseURL != tt.baseURL {
				t.Fatalf("BaseURL = %q, want %q", endpoint.BaseURL, tt.baseURL)
			}
		})
	}
}

func TestDockerHostEnv(t *testing.T) {
	tests := map[string]string{
		"/var/run/docker.sock":           "DOCKER_HOST=unix:///var/run/docker.sock",
		"tcp://docker.example:2375":      "DOCKER_HOST=tcp://docker.example:2375",
		`\\.\pipe\docker_engine`:         "DOCKER_HOST=npipe:////./pipe/docker_engine",
		"npipe:////./pipe/docker_engine": "DOCKER_HOST=npipe:////./pipe/docker_engine",
	}

	for input, want := range tests {
		if got := DockerHostEnv(input); got != want {
			t.Fatalf("DockerHostEnv(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestEndpointDialContextTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := ParseEndpoint("tcp://" + listener.Addr().String()).DialContext(ctx)
	if err != nil {
		t.Fatalf("DialContext tcp failed: %v", err)
	}
	conn.Close()

	select {
	case acceptedConn := <-accepted:
		acceptedConn.Close()
	case <-ctx.Done():
		t.Fatal("listener did not accept tcp connection")
	}
}

func TestEndpointDialContextErrorCases(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := ParseEndpoint("/tmp/hawser-missing.sock").DialContext(ctx); err == nil {
		t.Fatal("expected missing unix socket dial to fail")
	}

	_, err := ParseEndpoint(`\\.\pipe\hawser-missing`).DialContext(ctx)
	if runtime.GOOS == "windows" {
		if err == nil {
			t.Fatal("expected missing named pipe dial to fail")
		}
	} else if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected unsupported named pipe error, got %v", err)
	}
}
