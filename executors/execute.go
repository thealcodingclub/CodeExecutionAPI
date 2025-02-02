package executors

import (
	"CodeExecutionAPI/models"
	"bytes"
	"context"
	"errors"
	"os/exec"
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	defer cancel()

	start := time.Now()
	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}
	elapsed := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return models.ExecuteResponse{
			Output:     "",
			Error:      "Execution Timed Out",
			MemoryUsed: "0.0",
			CpuTime:    elapsed.String(),
		}, nil
	}

	if err != nil {
		return models.ExecuteResponse{
			Output:     "",
			Error:      stderr.String(),
			MemoryUsed: "0.0",
			CpuTime:    elapsed.String(),
		}, nil
	}

	return models.ExecuteResponse{
		Output:     out.String(),
		Error:      "",
		MemoryUsed: "0.0",
		CpuTime:    elapsed.String(),
	}, nil
}
