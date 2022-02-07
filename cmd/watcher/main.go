package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jonhanks/watcher"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Command struct {
	Action           []string
	WorkingDirectory string
}

func parseArguments() *watcher.Config {
	configPath := flag.String("config", ".watcher.json", "Config file path")
	flag.Parse()
	cfg := &watcher.Config{}
	f, err := os.Open(*configPath)
	if err != nil {
		log.Printf("Error opening config file. %v", err)
		os.Exit(1)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	if err = decoder.Decode(cfg); err != nil {
		log.Printf("Error parsing config file. %v", err)
		os.Exit(1)
	}
	return cfg
}

func main() {
	cfg := parseArguments()
	_ = cfg
	fmt.Println("Hello, starting to watch")
	if len(cfg.Monitor) == 0 {
		log.Fatalf("You must monitor at least one directory!")
	}

	done := make(chan bool)
	actions := make(chan Command)

	for i := 0; i < len(cfg.Monitor); i++ {
		go WatcherLoop(cfg.Monitor[i], actions)
	}
	go ActionLoop(actions)
	<-done
}

func IsExcluded(input string, exclusions []string) bool {
	if input == "" || len(exclusions) == 0 {
		return false
	}

	parts := strings.Split(input, string(os.PathSeparator))
	for _, part := range parts {
		for _, rule := range exclusions {
			if matched, err := filepath.Match(rule, part); err == nil && matched {
				return true
			}
		}
	}
	return false
}

func WatcherLoop(target watcher.Monitor, actions chan Command) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Unable to create filesystem watcher")
	}
	defer w.Close()
	err = w.Add(target.Directory)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if IsExcluded(event.Name, target.Exclude) {
				continue
			}
			log.Println("event: ", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
			}
			actions <- Command{
				Action:           target.Action,
				WorkingDirectory: target.Directory,
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func ActionLoop(actions chan Command) {
	for {
		select {
		case a := <-actions:
			runCommand(a)
		}
	}
}

func runCommand(c Command) {
	cmd := exec.Command(c.Action[0], c.Action[1:]...)
	cmd.Dir = c.WorkingDirectory
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	log.Println("Executing ", c.Action)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	io.Copy(os.Stdout, buf)
}
