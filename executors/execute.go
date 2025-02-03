package executors

import (
	"CodeExecutionAPI/models"
	"CodeExecutionAPI/resourcemanager"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func ExecuteCode(req models.ExecuteRequest) (models.ExecuteResponse, error) {
	var cmd *exec.Cmd
	maxMemoryFlag := fmt.Sprintf("--rlimit-as=%dk", req.MaxMemory)

	if !resourcemanager.ReserveMemory(req.MaxMemory) {
		return models.ExecuteResponse{
			Output:     "",
			Error:      "Resources Unavailable, Try again later",
			MemoryUsed: "0 KB",
			CpuTime:    "0.0s",
		}, errors.New("resources unavailable, try again later")
	}

	switch req.Language {
	case "python":
		cmd = exec.Command("firejail",
			"--private",
			"--quiet",
			"--noroot",
			"--caps.drop=all",
			"--read-only=/",
			"--net=none",
			maxMemoryFlag,
			"python3", "-c", req.Code)
	default:
		return models.ExecuteResponse{}, errors.New("unsupported language")
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Add context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			cmd.Process.Kill()
		}
	}()
	defer cancel()

	// Handle inputs
	if len(req.Inputs) > 0 {
		cmd.Stdin = strings.NewReader(strings.Join(req.Inputs, "\n"))
	}

	// Start execution timing
	start := time.Now()

	err := cmd.Start()
	if err != nil {
		return models.ExecuteResponse{
			Output:  "",
			Error:   fmt.Sprintf("Error starting command: %s", err),
			CpuTime: "0s",
		}, nil
	}

	// Wait for the process and capture resource usage
	var usage syscall.Rusage
	var waitStatus syscall.WaitStatus

	syscall.Wait4(cmd.Process.Pid, &waitStatus, 0, &usage)
	elapsed := time.Since(start)

	resourcemanager.ReleaseMemory(req.MaxMemory)

	memoryUsed := fmt.Sprintf("%d KB", usage.Maxrss)

	// Handle timeout separately
	if ctx.Err() == context.DeadlineExceeded {
		return models.ExecuteResponse{
			Output:     "",
			Error:      "Execution Timed Out",
			MemoryUsed: memoryUsed,
			CpuTime:    elapsed.String(),
		}, nil
	}

	return models.ExecuteResponse{
		Output:     out.String(),
		Error:      stderr.String(),
		MemoryUsed: memoryUsed,
		CpuTime:    elapsed.String(),
	}, nil
}
