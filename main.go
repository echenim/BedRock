package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func checkDay(ctx context.Context, dayOfTheWeek time.Weekday, hours int, workingDir string) {
	// Define file path relative to the working directory
	filePath := filepath.Join(workingDir, "example.txt")
	log.Printf("File path set to: %s", filePath)

	ticker := time.NewTicker(time.Duration(hours) * time.Hour)
	defer ticker.Stop()
	log.Printf("Timer started. Task will run every %d hours this %v.", hours, dayOfTheWeek.String())

	runCycle := func() {
		log.Println("Timer triggered. Starting task...")

		// Define content to overwrite in the file
		content := time.Now().Format(time.RFC1123) + "  ||   Hello, world! Hello, world! Hello, world!"

		// Step 1: Overwrite the file
		if err := updateFile(filePath, content); err != nil {
			log.Printf("Failed to update file: %v", err)
			return
		}

		// Step 2: Run Git commands
		if err := runGitCommands(filePath, workingDir); err != nil {
			log.Printf("Failed to execute Git commands: %v", err)
			return
		}

		log.Println("Task completed successfully.")
	}

	// Fire immediately on startup before entering ticker loop
	runCycle()
	if time.Now().Weekday() != dayOfTheWeek {
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received, stopping scheduler.")
			return
		case <-ticker.C:
			runCycle()
			if time.Now().Weekday() != dayOfTheWeek {
				return
			}
		}
	}
}

func updateFile(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		return err
	}
	log.Printf("File %s updated successfully.", filePath)
	return nil
}

func configureGitIdentity(workingDir string) error {
	name := os.Getenv("GIT_USER_NAME")
	email := os.Getenv("GIT_USER_EMAIL")
	if name == "" || email == "" {
		return fmt.Errorf("GIT_USER_NAME and GIT_USER_EMAIL environment variables must both be set")
	}

	for _, args := range [][]string{
		{"config", "--global", "user.name", name},
		{"config", "--global", "user.email", email},
	} {
		c := exec.Command("git", args...)
		c.Dir = workingDir
		var stderr bytes.Buffer
		c.Stderr = &stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("git %s: %s", args[0], stderr.String())
		}
	}

	log.Printf("Git identity set to: %s <%s>", name, email)
	return nil
}

func runGitCommands(filePath, workingDir string) error {
	if err := configureGitIdentity(workingDir); err != nil {
		return fmt.Errorf("failed to configure git identity: %w", err)
	}

	commands := []struct {
		name string
		args []string
	}{
		{"git", []string{"add", filePath}},
		{"git", []string{"commit", "-m", "updates:" + time.Now().Format(time.RFC1123)}},
		{"git", []string{"push"}},
	}

	for _, cmd := range commands {
		log.Printf("Executing: %s %v", cmd.name, cmd.args)
		c := exec.Command(cmd.name, cmd.args...)
		c.Dir = workingDir
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

	log.Println("Git commands executed successfully.")
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	schedule := map[time.Weekday]int{
		time.Monday:    7,
		time.Tuesday:   6,
		time.Wednesday: 4,
		time.Thursday:  8,
		time.Friday:    2,
		time.Saturday:  1,
		time.Sunday:    3,
	}

	for {
		currentDay := time.Now().Weekday()
		hours := schedule[currentDay]
		checkDay(ctx, currentDay, hours, workingDir)

		select {
		case <-ctx.Done():
			log.Println("Exiting.")
			return
		default:
		}
	}
}
