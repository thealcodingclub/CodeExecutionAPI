package executors

import (
	"CodeExecutionAPI/models"
	"CodeExecutionAPI/resourcemanager"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

func ExecuteCode(req models.ExecuteRequest) (models.ExecuteResponse, error) {
	var cmd *exec.Cmd
	maxMemoryFlag := fmt.Sprintf("--rlimit-as=%dk", req.MaxMemory)
	defer resourcemanager.ReleaseMemory(req.MaxMemory)

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
	case "c":
		processUUID := uuid.New().String()
		tmpFile := fmt.Sprintf("/tmp/%s.c", processUUID)
		binaryFile := fmt.Sprintf("/tmp/%s.c-out", processUUID)
		defer os.Remove(tmpFile)
		defer os.Remove(binaryFile)
		if err := os.WriteFile(tmpFile, []byte(req.Code), 0644); err != nil {
			return models.ExecuteResponse{}, errors.New("error writing temporary C file")
		}
		// Compile the C code
		compileCmd := exec.Command("gcc", tmpFile, "-o", binaryFile)
		var compileOut bytes.Buffer
		compileCmd.Stderr = &compileOut
		if err := compileCmd.Run(); err != nil {
			return models.ExecuteResponse{}, fmt.Errorf("compilation error: %v, %s", err, compileOut.String())
		}
		// Run the compiled binary in firejail
		cmd = exec.Command("firejail",
			"--private",
			"--quiet",
			"--noroot",
			"--caps.drop=all",
			"--read-only=/",
			"--net=none",
			maxMemoryFlag,
			binaryFile)

	case "cpp":
		processUUID := uuid.New().String()
		tmpFile := fmt.Sprintf("/tmp/%s.cpp", processUUID)
		binaryFile := fmt.Sprintf("/tmp/%s.cpp-out", processUUID)
		defer os.Remove(tmpFile)
		defer os.Remove(binaryFile)
		if err := os.WriteFile(tmpFile, []byte(req.Code), 0644); err != nil {
			return models.ExecuteResponse{}, errors.New("error writing temporary C++ file")
		}
		compileCmd := exec.Command("g++", tmpFile, "-o", binaryFile)
		var compileOut bytes.Buffer
		compileCmd.Stderr = &compileOut
		if err := compileCmd.Run(); err != nil {
			return models.ExecuteResponse{}, fmt.Errorf("compilation error: %v, %s", err, compileOut.String())
		}
		cmd = exec.Command("firejail",
			"--private",
			"--quiet",
			"--noroot",
			"--caps.drop=all",
			"--read-only=/",
			"--net=none",
			maxMemoryFlag,
			binaryFile)

	case "rust":
		processUUID := uuid.New().String()
		tmpFile := fmt.Sprintf("/tmp/%s.rs", processUUID)
		binaryFile := fmt.Sprintf("/tmp/%s.rs-out", processUUID)
		defer os.Remove(tmpFile)
		defer os.Remove(binaryFile)
		if err := os.WriteFile(tmpFile, []byte(req.Code), 0644); err != nil {
			return models.ExecuteResponse{}, errors.New("error writing Rust file")
		}
		compileCmd := exec.Command("rustc", tmpFile, "-o", binaryFile)
		var compileOut bytes.Buffer
		compileCmd.Stderr = &compileOut
		if err := compileCmd.Run(); err != nil {
			return models.ExecuteResponse{}, fmt.Errorf("compilation error: %v, %s", err, compileOut.String())
		}
		cmd = exec.Command("firejail",
			"--private",
			"--quiet",
			"--noroot",
			"--caps.drop=all",
			"--read-only=/",
			"--net=none",
			maxMemoryFlag,
			binaryFile)

	case "java":
		processUUID := uuid.New().String()
		tmpDir := fmt.Sprintf("/tmp/%s", processUUID)
		tmpFile := fmt.Sprintf("%s/Main.java", tmpDir)
		defer os.RemoveAll(tmpDir)
		if err := os.Mkdir(tmpDir, 0755); err != nil {
			return models.ExecuteResponse{}, errors.New("error creating temporary directory")
		}
		if err := os.WriteFile(tmpFile, []byte(req.Code), 0644); err != nil {
			return models.ExecuteResponse{}, errors.New("error writing Java file")
		}
		compileCmd := exec.Command("javac", tmpFile)
		var compileOut bytes.Buffer
		compileCmd.Stderr = &compileOut
		if err := compileCmd.Run(); err != nil {
			return models.ExecuteResponse{}, fmt.Errorf("compilation error: %v, %s", err, compileOut.String())
		}
		javaMemoryFlag := fmt.Sprintf("-Xmx%dk", req.MaxMemory)
		cmd = exec.Command("firejail",
			"--private",
			"--quiet",
			"--noroot",
			"--caps.drop=all",
			"--read-only=/",
			"--net=none",
			"java", javaMemoryFlag, "-cp", tmpDir, "Main")

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
	// var usage syscall.Rusage
	// var waitStatus syscall.WaitStatus
	// Potential problem in calling syscall.Wait4 directly instead of cmd.Wait(),
	// which bypasses the built-in process-wait logic in exec.Cmd and can lead
	// to missed exit codes or race conditions in the timeout goroutine
	// syscall.Wait4(cmd.Process.Pid, &waitStatus, 0, &usage)
	// elapsed := time.Since(start)
	errWait := cmd.Wait()
	elapsed := time.Since(start)
	var memoryUsed string
	if rusage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
		memoryUsed = fmt.Sprintf("%d KB", rusage.Maxrss)
	} else {
		memoryUsed = "Unknown"
	}

	// Handle command wait error
	if errWait != nil {
		return models.ExecuteResponse{
			Output:     out.String(),
			Error:      stderr.String(),
			MemoryUsed: memoryUsed,
			CpuTime:    elapsed.String(),
		}, nil
	}

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
