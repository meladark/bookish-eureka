package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}
	address := flag.Arg(0) + ":" + flag.Arg(1)
	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "...Connection error:", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "...Connected to", address)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		if err := client.Receive(); err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
			} else {
				fmt.Fprintln(os.Stderr, "...Receive error:", err)
			}
			close(done)
		}
	}()
	go func() {
		err := client.Send()
		if err != nil {
			if err.Error() == "connection closed by peer" {
				fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
			} else {
				fmt.Fprintln(os.Stderr, "...Send error:", err)
			}
		} else {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
		close(done)
	}()
	select {
	case <-sigCh:
		fmt.Fprintln(os.Stderr, "...SIGINT")
	case <-done:
	}
	client.Close()
}
