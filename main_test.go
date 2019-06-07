package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

var progPath string
var configPath = filepath.Join("example", "config", "config.yml")

const progName = "exporter_proxy"

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	var err error

	progPath, err = os.Getwd()
	if err != nil {
		fmt.Printf("can't get current dir :%s \n", err)
		os.Exit(1)
	}

	progPath = filepath.Join(progPath, progName)

	build := exec.Command("go", "build", "-o", progPath)
	output, err := build.CombinedOutput()
	if err != nil {
		fmt.Printf("compilation error: %s \n", output)
		os.Exit(1)
	}

	exitCode := m.Run()
	os.Remove(progPath)
	os.Exit(exitCode)
}

func TestStartupExitCodeWithInvalidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	fakeInputFile := "fake-input-file"
	expectedExitStatus := 1

	prog := exec.Command(progPath, "--config="+fakeInputFile)
	err := prog.Run()

	if err == nil {
		t.Errorf("Command execution succeeded with an invalid config file: %v", err)
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		status := exitError.Sys().(syscall.WaitStatus)
		if status.ExitStatus() != expectedExitStatus {
			t.Errorf("unexpected exit code with invalid configuration file")
		}
	} else {
		t.Errorf("unable to retrieve the exit status: %v", err)
	}
}

func TestStartupExitCodeWithValidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	prog := exec.Command(progPath, "--config="+configPath)
	err := prog.Start()

	if err != nil {
		t.Errorf("command execution error: %v", err)
		return
	}

	done := make(chan error)
	go func() {
		done <- prog.Wait()
	}()

	var startedOk bool
	var stoppedErr error

	for x := 0; x < 10; x++ {
		if _, err := http.Get("http://localhost:9099"); err == nil {

			startedOk = true
			prog.Process.Signal(os.Interrupt)
			select {
			case stoppedErr = <-done:
				break
			case <-time.After(10 * time.Second):
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !startedOk {
		t.Errorf("%s did not start in the specified timeout", progName)
		return
	}

	if err := prog.Process.Kill(); err == nil {
		t.Errorf("%s didn't shutdown gracefully after sending the Interrupt signal", progName)
	} else if stoppedErr != nil && stoppedErr.Error() != "signal: interrupt" {
		t.Errorf("%s exited with an unexpected error:%v", progName, stoppedErr)
	}
}
