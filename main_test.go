package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func newTestMonitor(read cpuReader, stdout, stderr io.Writer) cpuMonitor {
	return cpuMonitor{
		read:     read,
		procRoot: "/proc",
		nodeName: "node-a",
		stdout:   stdout,
		stderr:   stderr,
	}
}

func TestReadCPUStat(t *testing.T) {
	root := writeStatFile(t, "cpu  100 20 30 400 10 5 6 2 7 8\n")

	got, err := readCPUStat(root)
	if err != nil {
		t.Fatalf("readCPUStat() error = %v", err)
	}

	want := cpuStat{Idle: 410, Total: 573}
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

func TestReadCPUStatRejectsNonAggregateCPUFirstLine(t *testing.T) {
	root := writeStatFile(t, "cpu0  1 2 3 4 5 6 7 8\n")

	if _, err := readCPUStat(root); err == nil {
		t.Fatal("readCPUStat() error = nil, want aggregate CPU line error")
	}
}

func TestReadCPUStatReturnsMissingFileError(t *testing.T) {
	if _, err := readCPUStat(t.TempDir()); err == nil {
		t.Fatal("readCPUStat() error = nil, want missing file error")
	}
}

func TestCalculateCPUUsage(t *testing.T) {
	previous := cpuStat{Idle: 600, Total: 1000}
	current := cpuStat{Idle: 650, Total: 1200}

	got, err := calculateCPUUsage(previous, current)
	if err != nil {
		t.Fatalf("calculateCPUUsage() error = %v", err)
	}
	if math.Abs(got-75.0) > 0.001 {
		t.Fatalf("calculateCPUUsage() = %.3f, want 75.0", got)
	}
}

func TestCalculateCPUUsageRejectsZeroDelta(t *testing.T) {
	stat := cpuStat{Idle: 600, Total: 1000}
	if _, err := calculateCPUUsage(stat, stat); err == nil {
		t.Fatal("calculateCPUUsage() error = nil, want zero delta error")
	}
}

func TestCalculateCPUUsageRejectsDecreasingCounters(t *testing.T) {
	previous := cpuStat{Idle: 600, Total: 1000}
	current := cpuStat{Idle: 500, Total: 900}
	if _, err := calculateCPUUsage(previous, current); err == nil {
		t.Fatal("calculateCPUUsage() error = nil, want decreasing counter error")
	}
}

func TestCalculateCPUUsageRejectsIdleDeltaAboveTotalDelta(t *testing.T) {
	previous := cpuStat{Idle: 600, Total: 1000}
	current := cpuStat{Idle: 800, Total: 1100}
	if _, err := calculateCPUUsage(previous, current); err == nil {
		t.Fatal("calculateCPUUsage() error = nil, want invalid delta error")
	}
}

func TestMonitorReturnsInitialSampleError(t *testing.T) {
	wantErr := errors.New("read failed")
	reader := func(string) (cpuStat, error) {
		return cpuStat{}, wantErr
	}
	monitor := newTestMonitor(reader, &bytes.Buffer{}, &bytes.Buffer{})

	err := monitor.monitor(context.Background(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("monitor() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestMonitorWritesUsageForTick(t *testing.T) {
	stats := []cpuStat{
		{Idle: 600, Total: 1000},
		{Idle: 650, Total: 1200},
	}
	call := 0
	reader := func(string) (cpuStat, error) {
		stat := stats[call]
		call++
		return stat, nil
	}
	ticks := make(chan time.Time, 1)
	ticks <- time.Now()
	close(ticks)
	var stdout bytes.Buffer
	monitor := newTestMonitor(reader, &stdout, &bytes.Buffer{})

	if err := monitor.monitor(context.Background(), ticks); err != nil {
		t.Fatalf("monitor() error = %v", err)
	}
	if got, want := stdout.String(), "[Host: node-a] CPU: 75.0%\n"; got != want {
		t.Fatalf("monitor() output = %q, want %q", got, want)
	}
}

func TestMonitorStopsWhenContextIsCanceled(t *testing.T) {
	reader := func(string) (cpuStat, error) {
		return cpuStat{Idle: 600, Total: 1000}, nil
	}
	monitor := newTestMonitor(reader, &bytes.Buffer{}, &bytes.Buffer{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := monitor.monitor(ctx, nil); err != nil {
		t.Fatalf("monitor() error = %v, want nil", err)
	}
}

func TestMonitorStopsWhenTicksClose(t *testing.T) {
	reader := func(string) (cpuStat, error) {
		return cpuStat{Idle: 600, Total: 1000}, nil
	}
	ticks := make(chan time.Time)
	close(ticks)
	monitor := newTestMonitor(reader, &bytes.Buffer{}, &bytes.Buffer{})

	if err := monitor.monitor(context.Background(), ticks); err != nil {
		t.Fatalf("monitor() error = %v, want nil", err)
	}
}

func TestMonitorLogsSampleErrorAndContinues(t *testing.T) {
	readErr := errors.New("temporary read failure")
	results := []struct {
		stat cpuStat
		err  error
	}{
		{stat: cpuStat{Idle: 600, Total: 1000}},
		{err: readErr},
		{stat: cpuStat{Idle: 650, Total: 1200}},
	}
	call := 0
	reader := func(string) (cpuStat, error) {
		result := results[call]
		call++
		return result.stat, result.err
	}
	ticks := make(chan time.Time, 2)
	ticks <- time.Now()
	ticks <- time.Now()
	close(ticks)
	var stdout, stderr bytes.Buffer
	monitor := newTestMonitor(reader, &stdout, &stderr)

	if err := monitor.monitor(context.Background(), ticks); err != nil {
		t.Fatalf("monitor() error = %v", err)
	}
	if !strings.Contains(stderr.String(), readErr.Error()) {
		t.Fatalf("monitor() stderr = %q, want read error", stderr.String())
	}
	if got, want := stdout.String(), "[Host: node-a] CPU: 75.0%\n"; got != want {
		t.Fatalf("monitor() output = %q, want %q", got, want)
	}
}

func TestMonitorLogsCalculationErrorAndContinues(t *testing.T) {
	stats := []cpuStat{
		{Idle: 600, Total: 1000},
		{Idle: 600, Total: 1000},
		{Idle: 650, Total: 1200},
	}
	call := 0
	reader := func(string) (cpuStat, error) {
		stat := stats[call]
		call++
		return stat, nil
	}
	ticks := make(chan time.Time, 2)
	ticks <- time.Now()
	ticks <- time.Now()
	close(ticks)
	var stdout, stderr bytes.Buffer
	monitor := newTestMonitor(reader, &stdout, &stderr)

	if err := monitor.monitor(context.Background(), ticks); err != nil {
		t.Fatalf("monitor() error = %v", err)
	}
	if !strings.Contains(stderr.String(), "CPU usage calculation failed") {
		t.Fatalf("monitor() stderr = %q, want calculation error", stderr.String())
	}
	if got, want := stdout.String(), "[Host: node-a] CPU: 75.0%\n"; got != want {
		t.Fatalf("monitor() output = %q, want %q", got, want)
	}
}

func TestRunRejectsNonPositiveInterval(t *testing.T) {
	monitor := newTestMonitor(readCPUStat, &bytes.Buffer{}, &bytes.Buffer{})

	err := monitor.run(context.Background(), 0)
	if err == nil {
		t.Fatal("run() error = nil, want invalid interval error")
	}
}

func TestRunMainReturnsFailureWhenInitialSampleFails(t *testing.T) {
	t.Setenv("PROC_ROOT", t.TempDir())
	if got := runMain(&bytes.Buffer{}, &bytes.Buffer{}); got != 1 {
		t.Fatalf("runMain() = %d, want 1", got)
	}
}
