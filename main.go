package main

import (
	"CodeExecutionAPI/executors"
	"CodeExecutionAPI/models"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

func checkFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getLinuxDistro() string {
	if checkFileExists("/etc/arch-release") {
		return "Arch"
	} else if checkFileExists("/etc/debian_version") {
		return "Debian"
	} else if data, err := os.ReadFile("/etc/os-release"); err == nil {
		content := string(data)
		if strings.Contains(content, "Ubuntu") {
			return "Ubuntu"
		} else if strings.Contains(content, "Fedora") {
			return "Fedora"
		} else {
			return "Unknown"
		}
	} else {
		return "Unknown"
	}
}

func isCommandAvail(cmdName string) bool {
	Cmd := exec.Command("which", cmdName)
	err := Cmd.Run()
	return err == nil
}

func installFirejail(distro string) {
	if isCommandAvail("firejail") {
		fmt.Println("Firejail is already installed.")
	} else {
		var cmd *exec.Cmd

		switch distro {
		case "Debian", "Ubuntu":
			cmd = exec.Command("sudo", "apt", "install", "-y", "firejail")
		case "Arch":
			cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "firejail")
		case "Fedora":
			cmd = exec.Command("sudo", "dnf", "install", "-y", "firejail")
		default:
			fmt.Println("Unsupported distro, sorry.")
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatal("Error while installing Firejail.")
		} else {
			fmt.Println("Firejail insalled successfully.")
		}
	}
}

func main() {

	Os := runtime.GOOS

	if Os == "win32" {
		fmt.Println("You are on WINDOWS Ser, please use WSL")
		os.Exit(1)
	} else if Os == "linux" || Os == "linux-gnu" {
		distro := getLinuxDistro()

		if distro == "Unknown" {
			fmt.Println("Unknown Linux Distribution.")
		} else if distro == "Arch" || distro == "Debian" || distro == "Fedora" || distro == "Ubuntu" {
			installFirejail(distro)
		} else {
			fmt.Println("Unsupported OS ! ")
			os.Exit(1)
		}
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "https://github.com/thealcodingclub/CodeExecutionAPI")
	})

	r.GET("/execute", func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	r.POST("/execute", func(c *gin.Context) {
		var req models.ExecuteRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if req.Timeout == 0 {
			req.Timeout = 5
		}

		if req.Timeout > 60 {
			req.Timeout = 60
		}

		if req.MaxMemory == 0 {
			req.MaxMemory = 32768 // 32 MB
		}

		if req.MaxMemory > 131072 {
			req.MaxMemory = 131072 // 128 MB
		}

		// Execute the code
		output, err := executors.ExecuteCode(req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, output)
	})

	r.Run(":8080")
}
