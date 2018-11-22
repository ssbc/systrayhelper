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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestClicking(t *testing.T) {
	var err error
	r := require.New(t)
	logOut := os.Stderr

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

	var xvfb *exec.Cmd
	if _, ok := os.LookupEnv("TRAY_STARTXVFB"); ok {
		xvfb = exec.Command("Xvfb", ":23", "-screen", "0", "800x600x16")
		xvfb.Stdout = logOut
		xvfb.Stderr = logOut
		err = xvfb.Start()
		r.NoError(err, "failed to start virtual X framebuffer")
		fmt.Fprintln(logOut, "xvfb started. PID:", xvfb.Process.Pid)
		defer halt(xvfb)
	}

	if _, ok := os.LookupEnv("TRAY_WATCH"); ok {
		vncS := exec.Command("x11vnc", "-multiptr", "-display", ":23")
		vncS.Stdout = logOut
		vncS.Stderr = logOut
		vncS.Env = append(os.Environ(), "DISPLAY=:23")
		err = vncS.Start()
		r.NoError(err, "failed to start x11vnc")
		fmt.Fprintln(logOut, "vncS started. PID:", vncS.Process.Pid)
		defer halt(vncS)
		time.Sleep(time.Second * 10)
	}

	var ffmpeg *exec.Cmd
	if fname, ok := os.LookupEnv("TRAY_RECORD"); ok {
		ffmpeg = exec.Command("ffmpeg",
			"-y",
			"-an",
			"-f", "x11grab",
			"-framerate", "25",
			"-video_size", "cif",
			"-follow_mouse", "centered",
			"-draw_mouse", "1",
			"-i", ":23", fname)
		ffmpeg.Stderr = os.Stderr
		ffmpeg.Stdout = os.Stderr
		err = ffmpeg.Start()
		r.NoError(err, "failed to start ffmpeg recording")
		fmt.Fprintln(logOut, "ffmpeg started. PID:", ffmpeg.Process.Pid)
		time.Sleep(time.Second * 2)
	}

	i3 := exec.Command("i3", "-c", "i3_config")
	i3.Stdout = logOut
	i3.Stderr = logOut
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
	testJSON, err := os.Open("../test.json")
	r.NoError(err)
	th.Stdin = testJSON
	defer testJSON.Close()

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

	if ffmpeg != nil {
		ffmpeg.Process.Signal(os.Interrupt)
		err := ffmpeg.Wait()
		t.Log(err) // 255? look for "exited normaly"
	}

	if _, ok := os.LookupEnv("TRAY_STARTXVFB"); ok {
		t.Log("waiting for xvfb")
		err = xvfb.Wait()
		t.Log("xvfb err: err")
	}
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
