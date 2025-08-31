package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/cooper/quiki/webserver"
)

// handleSignals listens for SIGHUP and triggers a server config rehash.
func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for range c {
		log.Println("received SIGHUP: rehashing server config")
		if err := webserver.Rehash(); err != nil {
			log.Println("error rehashing server config:", err)
		}
	}
}

// handleReload sends SIGHUP to the running server
func handleReload() {
	// determine PID file path
	pidPath := pidFile
	if pidPath == "" {
		configPath := opts.Config
		if configPath == "" {
			configPath = filepath.Join(os.Getenv("HOME"), "quiki", "quiki.conf")
		}
		configDir := filepath.Dir(configPath)
		pidPath = filepath.Join(configDir, "server.pid")
	}

	// read PID from file
	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		log.Fatalf("failed to read PID file %s: %v", pidPath, err)
	}

	pidStr := strings.TrimSpace(string(pidData))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		log.Fatalf("invalid PID in file %s: %s", pidPath, pidStr)
	}

	// find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalf("failed to find process %d: %v", pid, err)
	}

	// send SIGHUP
	err = process.Signal(syscall.SIGHUP)
	if err != nil {
		log.Fatalf("failed to send SIGHUP to process %d: %v", pid, err)
	}

	log.Printf("sent reload signal to server (PID %d)", pid)
}

// writePIDFile writes the current process PID to the PID file
func writePIDFile() {
	// determine PID file path
	pidPath := pidFile
	if pidPath == "" {
		configPath := opts.Config
		if configPath == "" {
			configPath = filepath.Join(os.Getenv("HOME"), "quiki", "quiki.conf")
		}
		configDir := filepath.Dir(configPath)
		pidPath = filepath.Join(configDir, "server.pid")
	}

	// ensure directory exists
	if err := os.MkdirAll(filepath.Dir(pidPath), 0755); err != nil {
		log.Fatalf("failed to create PID file directory: %v", err)
	}

	// write PID to file
	pid := os.Getpid()
	err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d\n", pid)), 0644)
	if err != nil {
		log.Fatalf("failed to write PID file %s: %v", pidPath, err)
	}

	log.Printf("wrote PID %d to file %s", pid, pidPath)

	// set up cleanup on exit
	setupPIDCleanup(pidPath)
}

// setupPIDCleanup sets up signal handlers to clean up the PID file on exit
func setupPIDCleanup(pidPath string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("cleaning up PID file on exit")
		os.Remove(pidPath)
		os.Exit(0)
	}()
}
