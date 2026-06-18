package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultProcRoot = "/proc"
	defaultNodeName = "unknown"
	sampleInterval  = 5 * time.Second
	cpuFieldCount   = 8
	cpuIdleIndex    = 3
	cpuIOWaitIndex  = 4
)

type cpuStat struct {
	Idle  uint64
	Total uint64
}

type cpuReader func(string) (cpuStat, error)

type cpuMonitor struct {
	read     cpuReader
	procRoot string
	nodeName string
	stdout   io.Writer
	stderr   io.Writer
}

func readCPUStat(procRoot string) (cpuStat, error) {
	data, err := os.ReadFile(filepath.Join(procRoot, "stat"))
	if err != nil {
		return cpuStat{}, fmt.Errorf("read cpu stat: %w", err)
	}

	firstLine := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(firstLine)
	if len(fields) < 9 || fields[0] != "cpu" {
		return cpuStat{}, fmt.Errorf("invalid aggregate cpu line: %q", firstLine)
	}

	values := [cpuFieldCount]uint64{}
	for i := range values {
		value, err := strconv.ParseUint(fields[i+1], 10, 64)
		if err != nil {
			return cpuStat{}, fmt.Errorf("parse cpu field %d: %w", i+1, err)
		}
		values[i] = value
	}

	idle := values[cpuIdleIndex] + values[cpuIOWaitIndex]
	var total uint64
	for _, value := range values {
		total += value
	}

	return cpuStat{Idle: idle, Total: total}, nil
}

func calculateCPUUsage(previous, current cpuStat) (float64, error) {
	if current.Total <= previous.Total {
		return 0, fmt.Errorf("cpu total counter did not increase")
	}
	if current.Idle < previous.Idle {
		return 0, fmt.Errorf("cpu idle counter decreased")
	}

	totalDelta := current.Total - previous.Total
	idleDelta := current.Idle - previous.Idle
	if idleDelta > totalDelta {
		return 0, fmt.Errorf("cpu idle delta exceeds total delta")
	}

	return float64(totalDelta-idleDelta) / float64(totalDelta) * 100, nil
}

func envOrDefault(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

func formatCPUUsage(nodeName string, usage float64) string {
	return fmt.Sprintf("[Host: %s] CPU: %.1f%%", nodeName, usage)
}

func (m cpuMonitor) collect(previous cpuStat) cpuStat {
	current, err := m.read(m.procRoot)
	if err != nil {
		fmt.Fprintf(m.stderr, "CPU sample failed: %v\n", err)
		return previous
	}

	usage, err := calculateCPUUsage(previous, current)
	if err != nil {
		fmt.Fprintf(m.stderr, "CPU usage calculation failed: %v\n", err)
		return current
	}

	fmt.Fprintln(m.stdout, formatCPUUsage(m.nodeName, usage))
	return current
}

func (m cpuMonitor) monitor(ctx context.Context, ticks <-chan time.Time) error {
	previous, err := m.read(m.procRoot)
	if err != nil {
		return fmt.Errorf("initial CPU sample: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-ticks:
			if !ok {
				return nil
			}
			previous = m.collect(previous)
		}
	}
}

func (m cpuMonitor) run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("sample interval must be positive")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	return m.monitor(ctx, ticker.C)
}

func runMain(stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	monitor := cpuMonitor{
		read:     readCPUStat,
		procRoot: envOrDefault("PROC_ROOT", defaultProcRoot),
		nodeName: envOrDefault("NODE_NAME", defaultNodeName),
		stdout:   stdout,
		stderr:   stderr,
	}

	err := monitor.run(ctx, sampleInterval)
	if err != nil {
		fmt.Fprintf(stderr, "CPU monitor failed: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(runMain(os.Stdout, os.Stderr))
}
