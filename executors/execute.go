package executors

import (
	"CodeExecutionAPI/models"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

func ExecuteCode(req models.ExecuteRequest) (models.ExecuteResponse, error) {
	var cmd *exec.Cmd
	switch req.Language {
	case "python":
		cmd = exec.Command("python3", "-c", req.Code)
	case "c":
		cmd = exec.Command("sh", "-c", "echo '"+req.Code+"' | gcc -x c -o /tmp/a.out - && /tmp/a.out")
	case "java":
		cmd = exec.Command("sh", "-c", "echo '"+req.Code+"' > /tmp/Main.java && javac /tmp/Main.java && java -cp /tmp Main")
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

	_, waitErr := syscall.Wait4(cmd.Process.Pid, &waitStatus, 0, &usage)
	elapsed := time.Since(start)

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

	// Handle errors
	if waitErr != nil {
		return models.ExecuteResponse{
			Output:     "Error occured",
			Error:      stderr.String(),
			MemoryUsed: memoryUsed,
			CpuTime:    elapsed.String(),
		}, nil
	}

	// Return successful response
	return models.ExecuteResponse{
		Output:     out.String(),
		Error:      "",
		MemoryUsed: memoryUsed,
		CpuTime:    elapsed.String(),
	}, nil
}
