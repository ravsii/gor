package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
	"github.com/jessevdk/go-flags"
)

func main() {
	args, err := flags.NewParser(nil, flags.IgnoreUnknown).Parse()
	if err != nil {
		log.Fatal(err)
	}

	restartCh := make(chan struct{})

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)

				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}

				restartCh <- struct{}{}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Add a path.
	err = watcher.Add(cwd)
	if err != nil {
		log.Fatal(err)
	}

	var oldCmd *exec.Cmd

	go func() {
		for range restartCh {
			if oldCmd != nil {
				if err := oldCmd.Process.Kill(); err != nil {
					log.Println("can't kill old process, pid ", oldCmd.Process.Pid)
				}
			}

			cmd := exec.Command("go", args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			oldCmd = cmd

			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
		}
	}()

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// Block main goroutine forever.
	<-make(chan struct{})
}
