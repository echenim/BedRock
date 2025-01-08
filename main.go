package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Define file path relative to the main.go directory
	filePath := filepath.Join(workingDir, "example.txt")
	log.Printf("File path set to: %s", filePath)

	// Define a ticker that triggers every 4 hours
	// ticker := time.NewTicker(4 * time.Hour)
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	log.Println("Timer started. Task will run every 4 hours.")

	// Infinite loop to listen for ticker ticks
	for range ticker.C {
		log.Println("Timer triggered. Starting task...")

		// Define content to overwrite in the file
		content := time.Now().Format(time.RFC1123) + "  ||   Hello, world! Hello, world! Hello, world!"

		// Step 1: Overwrite the file
		if err := updateFile(filePath, content); err != nil {
			log.Printf("Failed to update file: %v", err)
			continue
		}

		// Step 2: Run Git commands
		if err := runGitCommands(); err != nil {
			log.Printf("Failed to execute Git commands: %v", err)
			continue
		}

		log.Println("Task completed successfully.")
	}
}

func updateFile(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		return err
	}
	log.Printf("\nFile %s updated successfully.\n", filePath)
	return nil
}

func runGitCommands() error {
	commands := []struct {
		name string
		args []string
	}{
		{"git", []string{"add", "."}},
		{"git", []string{"commit", "-m", "updates:" + time.Now().Format(time.RFC1123)}},
		{"git", []string{"push"}},
	}

	for _, cmd := range commands {
		log.Printf("Executing: %s %v", cmd.name, cmd.args)
		c := exec.Command(cmd.name, cmd.args...)
		var out bytes.Buffer
		var stderr bytes.Buffer
		c.Stdout = &out
		c.Stderr = &stderr

		err := c.Run()
		if err != nil {
			log.Printf("Error: %s", stderr.String())
			return err
		}
		log.Printf("Output: %s", out.String())
	}

	log.Println("\nGit commands executed successfully.")
	return nil
}
