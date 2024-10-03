package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

const (
	FORWARD_FILE  = "/proc/sys/net/ipv4/ip_forward"
	MAX_ERR_COUNT = 3
)

func main() {
	u, err := user.Current()
	if err != nil {
		fmt.Printf("error getting current user: %v\n", err)
		os.Exit(1)
	}

	if u.Uid != "0" {
		fmt.Printf("this program must be run as root\n")
		os.Exit(2)
	}

	value := flag.String("value", "1", "What value should ip_forward contain (0 or 1)")

	if *value != "1" && *value != "0" {
		fmt.Printf("invalid value %v to write to ip_forward\n", value)
		os.Exit(1)
	}

	err = checkAndFix(*value)
	if err != nil {
		fmt.Printf("error during initial file check: %v\n", err)
		os.Exit(1)
	}

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	exit := func(code int) {
		done()
		os.Exit(code)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("error creating watcher: %v\n", err)
		exit(1)
	}
	defer watcher.Close()

	err = watcher.Add(FORWARD_FILE)
	if err != nil {
		fmt.Printf("error adding forward file to watcher: %v\n", err)
		exit(1)
	}

	run := true
	errCount := 0
	for run {
		if errCount >= 3 {
			fmt.Printf("at least 3 errors in a row, exiting\n")
			exit(1)
		}
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				fmt.Printf("bad watcher event: %+v\n", event)
				errCount++
				continue
			}

			fmt.Printf("event: %+v\n", event)

			if event.Has(fsnotify.Write) {
				err = checkAndFix(*value)
				if err != nil {
					fmt.Printf("error writing value to file: %v\n", err)
					errCount++
					continue
				}

				errCount = 0
			}
		case err, ok := <-watcher.Errors:
			errCount++
			if !ok {
				fmt.Printf("bad watcher error: %v\n", err)
				continue
			}
			fmt.Printf("error from watcher: %v\n", err)
		case <-ctx.Done():
			run = false
		}
	}

	fmt.Println("tetelestai")
	exit(0)
}

func checkAndFix(expected string) error {
	content, err := os.ReadFile(FORWARD_FILE)
	if err != nil {
		return fmt.Errorf("error reading ip_forward file: %w", err)
	}
	fmt.Printf("forward file contains: %v\n", strings.TrimSpace(string(content)))

	if strings.TrimSpace(string(content)) != expected {
		fmt.Printf("overwriting forward file with: %v\n", expected)
		err := os.WriteFile(FORWARD_FILE, []byte(expected), 0644)
		if err != nil {
			return fmt.Errorf("error writing ip_forward file: %w", err)
		}
	}

	return nil
}
