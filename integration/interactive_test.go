package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClicking(t *testing.T) {
	var err error
	r := require.New(t)

	var neededTools = []string{
		"Xvfb",         // virtual X11 server
		"i3", "i3-msg", // for known bar location (lower right area)
		"xdotool", // simulate clicks
		"systrayhelper"}
	for _, n := range neededTools {
		location, err := exec.LookPath(n)
		r.NoError(err, "did not find %s", n)
		t.Logf("found %s here: %s", n, location)
	}

	xvfb := exec.Command("Xvfb", ":23", "-screen", "0", "800x600x16")
	xvfb.Stderr = os.Stderr
	err = xvfb.Start()
	r.NoError(err, "failed to start virtual X framebuffer")
	fmt.Println("xvfb started. PID:", xvfb.Process.Pid)
	defer halt(xvfb)

	i3 := exec.Command("i3")
	i3.Stdout = os.Stderr
	i3.Stderr = os.Stderr
	i3.Env = append(os.Environ(), "DISPLAY=:23")
	err = i3.Start()
	r.NoError(err, "failed to start i3 (for its tray area)")
	fmt.Println("i3 started. PID:", i3.Process.Pid)
	defer halt(i3)

	th := exec.Command("systrayhelper")
	th.Stderr = os.Stderr
	th.Env = append(os.Environ(), "DISPLAY=:23")

	testJson, err := os.Open("../test.json")
	r.NoError(err)
	th.Stdin = testJson
	defer testJson.Close()

	stdout, err := th.StdoutPipe()
	r.NoError(err)

	err = th.Start()
	r.NoError(err, "failed to start the actual helper")
	fmt.Println("helper started. PID:", th.Process.Pid)
	defer halt(th)

	go func() {
		dec := json.NewDecoder(stdout)
		for {
			var v map[string]interface{}
			err = dec.Decode(&v)
			if err == io.EOF {
				break
			}
			check(err)
			t.Log(v)
		}
	}()

	fmt.Println("waiting for trayhelper")
	err = th.Wait()

	i3exit := exec.Command("i3-msg", "exit")
	i3exit.Stdout = os.Stderr
	i3exit.Stderr = os.Stderr
	i3exit.Env = append(os.Environ(), "DISPLAY=:23")
	out, err := i3exit.CombinedOutput()
	r.NoError(err, "failed to start the shut down i3: %s", string(out))

	fmt.Println("waiting for i3")
	err = i3.Wait()
	r.NoError(err)

	fmt.Println("waiting for xvfb")
	err = xvfb.Wait()
	r.NoError(err)
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "go routine check failed: %#v\n", err)
		debug.PrintStack()
		os.Exit(1)
	}
}

func halt(cmd *exec.Cmd) {
	fmt.Fprintln(os.Stderr, "halting:", cmd.Path)
	cmd.Process.Kill()
}
