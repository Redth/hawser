package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeCommandEnvSetsDockerHost(t *testing.T) {
	t.Setenv("DOCKER_HOST", "tcp://old.example:2375")

	client := NewComposeClient(`\\.\pipe\docker_engine`, t.TempDir())
	env := client.commandEnv([]string{"EXTRA=value"})

	var dockerHostValues []string
	for _, entry := range env {
		if strings.HasPrefix(entry, "DOCKER_HOST=") {
			dockerHostValues = append(dockerHostValues, entry)
		}
	}
	if len(dockerHostValues) != 1 {
		t.Fatalf("expected one DOCKER_HOST entry, got %v", dockerHostValues)
	}
	if dockerHostValues[0] != "DOCKER_HOST=npipe:////./pipe/docker_engine" {
		t.Fatalf("DOCKER_HOST = %q", dockerHostValues[0])
	}
	if env[len(env)-1] != "EXTRA=value" {
		t.Fatalf("extra env was not appended, got final entry %q", env[len(env)-1])
	}
}

func TestIsPathWithinBase(t *testing.T) {
	base := t.TempDir()
	inside := filepath.Join(base, "stack", "compose.yml")
	outside := filepath.Join(os.TempDir(), "hawser-outside-compose.yml")

	if !isPathWithinBase(base, base) {
		t.Fatal("base directory should be inside itself")
	}
	if !isPathWithinBase(base, inside) {
		t.Fatalf("%s should be within %s", inside, base)
	}
	if isPathWithinBase(base, outside) {
		t.Fatalf("%s should not be within %s", outside, base)
	}
	if isPathWithinBase(base, filepath.Join(base, "..", filepath.Base(outside))) {
		t.Fatal("path traversal should not be within base")
	}
}
