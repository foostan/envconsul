package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	var errExit bool
	var reload string
	var timeout time.Duration
	var consulAddr string
	var consulDC string
	var sanitize bool
	var upcase bool
	flag.Usage = usage
	flag.BoolVar(
		&errExit, "errexit", false,
		"exit if there is an error watching config keys")
	flag.DurationVar(
		&timeout, "timeout", 3*time.Second,
		"how long to wait after SIGTERM when reloading")
	flag.StringVar(
		&reload, "reload", "false",
		`if true, restarts the process when config change
			if terminate, kills the process when config change`)
	flag.StringVar(
		&consulAddr, "addr", "127.0.0.1:8500",
		"consul HTTP API address with port")
	flag.StringVar(
		&consulDC, "dc", "",
		"consul datacenter, uses local if blank")
	flag.BoolVar(
		&sanitize, "sanitize", true,
		"turn invalid characters in the key into underscores")
	flag.BoolVar(
		&upcase, "upcase", true,
		"make all environmental variable keys uppercase")
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		return 1
	}
	reloadOpts := map[string]bool{
		"true":      true,
		"false":     true,
		"terminate": true,
	}
	if !reloadOpts[reload] {
		fmt.Println("Invalid value for -reload. Possible values are true, false, and terminate")
		flag.Usage()
		return 111
	}

	args := flag.Args()
	config := WatchConfig{
		ConsulAddr: consulAddr,
		ConsulDC:   consulDC,
		Cmd:        args[1:],
		ErrExit:    errExit,
		Prefix:     args[0],
		Reload:     reload,
		Timeout:    timeout,
		Sanitize:   sanitize,
		Upcase:     upcase,
	}
	result, err := watchAndExec(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 111
	}

	return result
}

func usage() {
	cmd := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, strings.TrimSpace(helpText)+"\n\n", cmd)
	flag.PrintDefaults()
}

const helpText = `
Usage: %s [options] prefix child...

  Sets environmental variables for the child process by reading
  K/V from Consul's K/V store with the given prefix.

Options:
`
