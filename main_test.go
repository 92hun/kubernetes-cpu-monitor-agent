package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("PROC_ROOT", "/custom/proc")
	if got := envOrDefault("PROC_ROOT", "/proc"); got != "/custom/proc" {
		t.Fatalf("envOrDefault() = %q, want %q", got, "/custom/proc")
	}

	t.Setenv("PROC_ROOT", "")
	if got := envOrDefault("PROC_ROOT", "/proc"); got != "/proc" {
		t.Fatalf("envOrDefault() = %q, want default %q", got, "/proc")
	}
}

func TestFormatCPUUsage(t *testing.T) {
	got := formatCPUUsage("minikube", 12.54)
	want := "[Host: minikube] CPU: 12.5%"
	if got != want {
		t.Fatalf("formatCPUUsage() = %q, want %q", got, want)
	}
}

func writeStatFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "stat"), []byte(content), 0o600); err != nil {
		t.Fatalf("write stat fixture: %v", err)
	}
	return dir
}

func TestReadCPUStat(t *testing.T) {
	root := writeStatFile(t, "cpu  100 20 30 400 10 5 6 2 7 8\n")

	got, err := readCPUStat(root)
	if err != nil {
		t.Fatalf("readCPUStat() error = %v", err)
	}

	want := CPUStat{Idle: 410, Total: 573}
	if got != want {
		t.Fatalf("readCPUStat() = %+v, want %+v", got, want)
	}
}

func TestReadCPUStatRejectsTooFewFields(t *testing.T) {
	root := writeStatFile(t, "cpu  1 2 3\n")

	if _, err := readCPUStat(root); err == nil {
		t.Fatal("readCPUStat() error = nil, want field count error")
	}
}

func TestReadCPUStatRejectsInvalidNumber(t *testing.T) {
	root := writeStatFile(t, "cpu  1 2 invalid 4 5 6 7 8\n")

	if _, err := readCPUStat(root); err == nil {
		t.Fatal("readCPUStat() error = nil, want parse error")
	}
}

func TestCalculateCPUUsage(t *testing.T) {
	previous := CPUStat{Idle: 600, Total: 1000}
	current := CPUStat{Idle: 650, Total: 1200}

	got, err := calculateCPUUsage(previous, current)
	if err != nil {
		t.Fatalf("calculateCPUUsage() error = %v", err)
	}
	if math.Abs(got-75.0) > 0.001 {
		t.Fatalf("calculateCPUUsage() = %.3f, want 75.0", got)
	}
}

func TestCalculateCPUUsageRejectsZeroDelta(t *testing.T) {
	stat := CPUStat{Idle: 600, Total: 1000}
	if _, err := calculateCPUUsage(stat, stat); err == nil {
		t.Fatal("calculateCPUUsage() error = nil, want zero delta error")
	}
}

func TestCalculateCPUUsageRejectsDecreasingCounters(t *testing.T) {
	previous := CPUStat{Idle: 600, Total: 1000}
	current := CPUStat{Idle: 500, Total: 900}
	if _, err := calculateCPUUsage(previous, current); err == nil {
		t.Fatal("calculateCPUUsage() error = nil, want decreasing counter error")
	}
}

func TestCalculateCPUUsageRejectsIdleDeltaAboveTotalDelta(t *testing.T) {
	previous := CPUStat{Idle: 600, Total: 1000}
	current := CPUStat{Idle: 800, Total: 1100}
	if _, err := calculateCPUUsage(previous, current); err == nil {
		t.Fatal("calculateCPUUsage() error = nil, want invalid delta error")
	}
}
