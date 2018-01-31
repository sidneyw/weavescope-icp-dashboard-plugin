package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// We put the socket in a sub-directory to have more control on the permissions
	const socketPath = "/var/run/scope/plugins/icp-dashboard/icp-dashboard.sock"
	hostID, _ := os.Hostname()

	fmt.Printf("Current Host IP %s\n", GetOutboundIP())

	// Handle the exit signal
	setupSignals(socketPath)

	log.Printf("Starting on %s...\n", hostID)

	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	var plugin *Plugin = &Plugin{
		ID:          "icp-dashboard",
		Label:       "ICP Dashboard",
		Description: "Links into the ICP Dashboard",
		Interfaces:  []string{"reporter"},
		APIVersion:  1,

		Controls: Controls{false, false},
	}

	http.HandleFunc("/report", plugin.HandleReport)
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}

	// rpt := plugin.GenerateReport()

	// raw, err := json.Marshal(&rpt)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// jsonIndented, _ := prettyprint(raw)
	// fmt.Printf("%s", jsonIndented)
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func setupSocket(socketPath string) (net.Listener, error) {
	os.RemoveAll(filepath.Dir(socketPath))
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", filepath.Dir(socketPath), err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %v", socketPath, err)
	}

	log.Printf("Listening on: unix://%s", socketPath)
	return listener, nil
}

func setupSignals(socketPath string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		os.RemoveAll(filepath.Dir(socketPath))
		os.Exit(0)
	}()
}
