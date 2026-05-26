package docker

import "testing"

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
