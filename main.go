package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type CPUStat struct {
	Idle  uint64
	Total uint64
}

func readCPUStat(procRoot string) (CPUStat, error) {
	data, err := os.ReadFile(filepath.Join(procRoot, "stat"))
	if err != nil {
		return CPUStat{}, fmt.Errorf("read cpu stat: %w", err)
	}

	firstLine := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(firstLine)
	if len(fields) < 9 || fields[0] != "cpu" {
		return CPUStat{}, fmt.Errorf("invalid aggregate cpu line: %q", firstLine)
	}

	values := make([]uint64, 8)
	for i := range values {
		value, err := strconv.ParseUint(fields[i+1], 10, 64)
		if err != nil {
			return CPUStat{}, fmt.Errorf("parse cpu field %d: %w", i+1, err)
		}
		values[i] = value
	}

	idle := values[3] + values[4]
	var total uint64
	for _, value := range values {
		total += value
	}

	return CPUStat{Idle: idle, Total: total}, nil
}

func calculateCPUUsage(previous, current CPUStat) (float64, error) {
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

func main() {
	procRoot := envOrDefault("PROC_ROOT", "/proc")
	nodeName := envOrDefault("NODE_NAME", "unknown")

	previous, err := readCPUStat(procRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initial CPU sample failed: %v\n", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current, err := readCPUStat(procRoot)
			if err != nil {
				fmt.Fprintf(os.Stderr, "CPU sample failed: %v\n", err)
				continue
			}

			usage, err := calculateCPUUsage(previous, current)
			previous = current
			if err != nil {
				fmt.Fprintf(os.Stderr, "CPU usage calculation failed: %v\n", err)
				continue
			}

			fmt.Println(formatCPUUsage(nodeName, usage))
		}
	}
}
