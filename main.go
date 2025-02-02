package main

import (
	"CodeExecutionAPI/executors"
	"CodeExecutionAPI/models"

	"github.com/gin-gonic/gin"
)

// Main function to start the server
func main() {
	r := gin.Default()

	r.POST("/execute", func(c *gin.Context) {
		var req models.ExecuteRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if req.Timeout == 0 {
			req.Timeout = 5
		}

		// Execute the code
		output, err := executors.ExecuteCode(req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, output)
	})

	// Start the server
	r.Run(":8080") // Default port is 8080
}
