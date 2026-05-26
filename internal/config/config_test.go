package config

import (
	"testing"

	"github.com/Finsys/hawser/internal/docker"
)

func TestPlatformDefaults(t *testing.T) {
	if got := detectDockerSocketForGOOS("windows", ""); got != docker.DefaultNpipeSocket {
		t.Fatalf("windows Docker socket = %q, want %q", got, docker.DefaultNpipeSocket)
	}
	if got := detectDockerSocketForGOOS("linux", "/home/test"); got != docker.DefaultUnixSocket {
		t.Fatalf("linux Docker socket fallback = %q, want %q", got, docker.DefaultUnixSocket)
	}
	if got := defaultStacksDirForGOOS("windows"); got != `C:\ProgramData\hawser\stacks` {
		t.Fatalf("windows stacks dir = %q", got)
	}
	if got := defaultStacksDirForGOOS("linux"); got != "/data/stacks" {
		t.Fatalf("linux stacks dir = %q", got)
	}
}

func TestGetDockerEndpoint(t *testing.T) {
	cfg := &Config{DockerSocket: `\\.\pipe\docker_engine`}
	if got := cfg.GetDockerEndpoint(); got != "npipe:////./pipe/docker_engine" {
		t.Fatalf("npipe endpoint = %q", got)
	}

	cfg = &Config{DockerSocket: "/var/run/docker.sock"}
	if got := cfg.GetDockerEndpoint(); got != "unix:///var/run/docker.sock" {
		t.Fatalf("unix endpoint = %q", got)
	}

	cfg = &Config{DockerSocket: "/var/run/docker.sock", DockerHost: "tcp://localhost:2375"}
	if got := cfg.GetDockerEndpoint(); got != "tcp://localhost:2375" {
		t.Fatalf("DockerHost override endpoint = %q", got)
	}
}
