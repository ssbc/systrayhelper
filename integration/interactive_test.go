package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"testing"
	"time"

	"github.com/cryptix/go/logging/logtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestClicking(t *testing.T) {
	var err error
	r := require.New(t)
	//logOut:= os.Stderr
	logOut := logtest.Logger("Xlog", t)

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
	xvfb.Stderr = logOut
	err = xvfb.Start()
	r.NoError(err, "failed to start virtual X framebuffer")
	fmt.Fprintln(logOut, "xvfb started. PID:", xvfb.Process.Pid)
	defer halt(xvfb)

	if _, ok := os.LookupEnv("TRAY_DEBUG"); ok {
		vncS := exec.Command("x11vnc", "-multiptr", "-display", ":23")
		vncS.Stdout = logOut
		vncS.Stderr = logOut
		vncS.Env = append(os.Environ(), "DISPLAY=:23")
		err = vncS.Start()
		r.NoError(err, "failed to start vncS (for its tray area)")
		fmt.Println("vncS started. PID:", vncS.Process.Pid)
		defer halt(vncS)

		time.Sleep(time.Second * 5)
	}

	i3 := exec.Command("i3")
	//i3.Stdout = logOut
	//i3.Stderr = logOut
	i3.Env = append(os.Environ(), "DISPLAY=:23")
	err = i3.Start()
	r.NoError(err, "failed to start i3 (for its tray area)")
	fmt.Fprintln(logOut, "i3 started. PID:", i3.Process.Pid)
	defer halt(i3)

	th := exec.Command("systrayhelper")
	th.Stderr = logOut
	th.Env = append(os.Environ(), "DISPLAY=:23")
	stdout, err := th.StdoutPipe()
	r.NoError(err)
	testJson, err := os.Open("../test.json")
	r.NoError(err)
	th.Stdin = testJson
	defer testJson.Close()

	err = th.Start()
	r.NoError(err, "failed to start the actual helper")
	fmt.Fprintln(logOut, "helper started. PID:", th.Process.Pid)
	defer halt(th)

	thready := make(chan struct{})
	clickSent := make(chan struct{})
	go func() {
		var clicked bool
		dec := json.NewDecoder(stdout)
		for {
			var v map[string]interface{}
			err = dec.Decode(&v)
			if err == io.EOF {
				break
			}
			check(err)
			fmt.Fprintf(logOut, "got stdout (c?:%v): %+v\n", clicked, v)

			if is, ok := v["type"]; ok && is == "ready" {
				close(thready)
				<-clickSent
				clicked = true
				continue
			}

			if is, ok := v["type"]; ok && clicked && is == "clicked" {
				halt(th)
			}
		}
	}()

	go func() {
		<-thready

		xdt := exec.Command("xdotool",
			"mousemove", "793", "593",
			"sleep", "1",
			"click", "1",
			"sleep", "1",
			"mousemove_relative", "--", "0", "-20",
			"sleep", "1",
			"click", "1")
		xdt.Env = append(os.Environ(), "DISPLAY=:23")
		out, err := xdt.CombinedOutput()
		check(errors.Wrapf(err, "failed to click menu: %s", string(out)))

		fmt.Fprintln(logOut, "clicks send")
		close(clickSent)
	}()

	fmt.Fprintln(logOut, "waiting for trayhelper")
	err = th.Wait()

	i3exit := exec.Command("i3-msg", "exit")
	i3exit.Env = append(os.Environ(), "DISPLAY=:23")
	i3exit.Stdout = logOut
	i3exit.Stderr = logOut
	err = i3exit.Run()
	t.Log("i3-msg:", err)
	//r.NoError(err, "failed to start the shut down i3")

	t.Log("waiting for i3")
	err = i3.Wait()
	r.NoError(err)

	t.Log("waiting for xvfb")
	xvfb.Wait()
	//err = xvfb.Wait()
	//r.NoError(err)
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
