package main

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

type ExecuteRequest struct {
	Language    string `json:"language"`
	Code        string `json:"code"`
	Timeout     int    `json:"timeout,omitempty"`
	MemoryLimit int    `json:"memory_limit,omitempty"`
}

type ExecuteResponse struct {
	Output     string `json:"output"`
	Error      string `json:"error"`
	MemoryUsed string `json:"memory_used"`
	CpuTime    string `json:"cpu_time"`
}

func main() {
	r := gin.Default()

	// Define a simple route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Gin!",
		})
	})

	// Define the /execute route
	r.POST("/execute", func(c *gin.Context) {
		var req ExecuteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// Set default timeout if not provided
		if req.Timeout == 0 {
			req.Timeout = 5
		}

		// Execute the code
		output, err := executeCode(req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, output)
	})

	// Start the server
	r.Run(":8080") // Default port is 8080
}

func executeCode(req ExecuteRequest) (ExecuteResponse, error) {
	var cmd *exec.Cmd
	switch req.Language {
	case "python":
		cmd = exec.Command("python3", "-c", req.Code)
	case "c":
		cmd = exec.Command("sh", "-c", "echo '"+req.Code+"' | gcc -x c -o /tmp/a.out - && /tmp/a.out")
	case "java":
		cmd = exec.Command("sh", "-c", "echo '"+req.Code+"' > /tmp/Main.java && javac /tmp/Main.java && java -cp /tmp Main")
	default:
		return ExecuteResponse{}, errors.New("unsupported language")
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
		return ExecuteResponse{
			Output:     "",
			Error:      "Execution Timed Out",
			MemoryUsed: "0.0",
			CpuTime:    elapsed.String(),
		}, nil
	}

	if err != nil {
		return ExecuteResponse{
			Output:     "",
			Error:      stderr.String(),
			MemoryUsed: "0.0",
			CpuTime:    elapsed.String(),
		}, nil
	}

	return ExecuteResponse{
		Output:     out.String(),
		Error:      "",
		MemoryUsed: "0.0",
		CpuTime:    elapsed.String(),
	}, nil
}
