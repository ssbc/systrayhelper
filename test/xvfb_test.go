package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
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
		//"Xvfb",         // virtual X11 server
		"i3", "i3-msg", // for known bar location (lower right area)
		"xdotool", // simulate clicks
		"systrayhelper"}
	for _, n := range neededTools {
		location, err := exec.LookPath(n)
		r.NoError(err, "did not find %s", n)
		t.Logf("found %s here: %s", n, location)
	}

	var xvfb *exec.Cmd
	disp, ok := os.LookupEnv("DISPLAY")
	if !ok {
		t.Fatal("sorry - need a DISPLAY. maybe try xvfb-run")
		// todo: need to export DISPLAY back
		// using xvfb_run around 'go test' seems nicer
		xvfb = exec.Command("Xvfb", ":23", "-screen", "0", "800x600x16")
		xvfb.Stdout = logOut
		xvfb.Stderr = logOut
		err = xvfb.Start()
		r.NoError(err, "failed to start virtual X framebuffer")
		fmt.Fprintln(logOut, "xvfb started. PID:", xvfb.Process.Pid)
		defer halt(xvfb)
		disp = ":23"
	}

	xdt := exec.Command("xdotool", "getdisplaygeometry")
	out, err := xdt.CombinedOutput()
	r.NoError(errors.Wrapf(err, "failed to click menu: %s", string(out)))

	xy := strings.Split(strings.TrimSpace(string(out)), " ")
	r.Len(xy, 2)
	dispx, err := strconv.Atoi(xy[0])
	r.NoError(errors.Wrapf(err, "failed to decode X: %s", xy))
	dispy, err := strconv.Atoi(xy[1])
	r.NoError(errors.Wrapf(err, "failed to decode y: %s", xy))

	thx := fmt.Sprint(dispx - 7)
	thy := fmt.Sprint(dispy - 7)

	if _, ok := os.LookupEnv("TRAY_WATCH"); ok {
		vncS := exec.Command("x11vnc", "-multiptr", "-display", disp)
		vncS.Stdout = logOut
		vncS.Stderr = logOut
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
			"-follow_mouse", "centered",
			"-draw_mouse", "1",
			"-framerate", "10",
			"-video_size", "cif",
			"-i", disp,
			fname)
		ffmpeg.Stderr = os.Stderr
		ffmpeg.Stdout = os.Stderr
		err = ffmpeg.Start()
		r.NoError(err, "failed to start ffmpeg recording")
		fmt.Fprintln(logOut, "ffmpeg started. PID:", ffmpeg.Process.Pid)
		time.Sleep(time.Second * 2)
	}

	var i3 *exec.Cmd
	if _, ok := os.LookupEnv("TRAY_I3"); ok {
		i3 = exec.Command("i3", "-V", "-a", "-c", "i3_config")
		i3.Stdout = logOut
		i3.Stderr = logOut
		err = i3.Start()
		r.NoError(err, "failed to start i3 (for its tray area)")
		fmt.Fprintln(logOut, "i3 started. PID:", i3.Process.Pid)
		defer halt(i3)
	}

	th := exec.Command("systrayhelper")
	th.Stderr = logOut
	stdout, err := th.StdoutPipe()
	r.NoError(err)
	stdin, err := th.StdinPipe()
	r.NoError(err)

	err = th.Start()
	r.NoError(err, "failed to start the actual helper")
	fmt.Fprintln(logOut, "helper started. PID:", th.Process.Pid)
	defer halt(th)

	testJSON, err := os.Open("../test.json")
	r.NoError(err)

	n, err := io.Copy(stdin, testJSON)
	r.NoError(err)
	testJSON.Close()

	t.Logf("sent %d bytes of init json", n)

	msgs := make(chan interface{})
	go func() {
		enc := json.NewEncoder(stdin)
		for m := range msgs {
			err = enc.Encode(m)
			if err == io.EOF {
				break
			}
			check(errors.Wrap(err, "encode error to helper"))
			fmt.Fprintf(logOut, "sent msg %+v\n", m)
		}
	}()

	thready := make(chan struct{})
	startLvl1 := make(chan struct{})
	startLvl2 := make(chan struct{})
	startLvl3 := make(chan struct{})
	startLvl4 := make(chan struct{})
	lvl := 0
	go func() {
		dec := json.NewDecoder(stdout)
		for {
			var v Action
			err = dec.Decode(&v)
			if err == io.EOF {
				break
			}
			check(errors.Wrap(err, "decode error from helper"))
			fmt.Fprintf(logOut, "lvl:%d - %+v\n", lvl, v)

			if v.Type == "ready" {
				close(thready)
				<-startLvl1
				lvl++
				continue
			}

			if v.Type == "clicked" {
				t.Logf("lvl:%d - clicked %d: %s", lvl, v.SeqID, v.Item.Title)
				switch {
				case lvl == 1 && v.Item.Title == "check new title":
					close(startLvl2)
				case lvl == 2 && v.SeqID == 0 && v.Item.Title == "enabled":
					close(startLvl3)
				case lvl == 3 && v.SeqID == 0 && v.Item.Title == "now with more items":
					close(startLvl4)
				case lvl == 4 && v.SeqID == 2 && v.Item.Title == "quit":
					stdin.Close()
				default:
					t.Log("unhandled case", lvl, v)
				}
			}
		}
	}()

	// level 1
	go func() {
		<-thready

		msgs <- Action{
			Type: "update-item", SeqID: 555,
			Item: Item{Title: "does not exist"},
		}

		xdt := exec.Command("xdotool",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"sleep", "1",
			"mousemove_relative", "--", "0", "-20",
			"sleep", "1",
			"click", "1")
		out, err := xdt.CombinedOutput()
		check(errors.Wrapf(err, "failed to click menu: %s", string(out)))

		fmt.Fprintln(logOut, "clicks send")
		close(startLvl1)
	}()

	go func() { // level 2
		<-startLvl2
		lvl++
		msgs <- Action{
			Type: "update-item", SeqID: 0,
			Item: Item{Title: "enabled", Enabled: true},
		}

		msgs <- Action{
			Type: "update-item", SeqID: 1,
			Item: Item{Title: "hidden", Hidden: true},
		}

		xdt := exec.Command("xdotool",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"sleep", "1",
			"mousemove_relative", "--", "0", "-50",
			"sleep", "1",
			"click", "1")
		out, err := xdt.CombinedOutput()
		check(errors.Wrapf(err, "failed to click menu: %s", string(out)))

		fmt.Fprintln(logOut, "level2 send")
	}()

	go func() { // level 3
		<-startLvl3
		lvl++
		msgs <- Action{
			Type: "append-item",
			Item: Item{Title: "appended", Enabled: true},
		}
		msgs <- Action{
			Type: "update-item", SeqID: 0,
			Item: Item{Title: "now with more items", Enabled: true},
		}

		xdt := exec.Command("xdotool",
			// siigh... somehow the drawing is buggy sometimes
			"mousemove", "0", "0", "click", "1",
			"mousemove", thx, thy, "click", "1",
			"sleep", "1",
			"mousemove", "0", "0", "click", "1",
			"sleep", "1",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"sleep", "1",
			"mousemove_relative", "--", "0", "-65",
			"sleep", "1",
			"click", "1")
		out, err := xdt.CombinedOutput()
		check(errors.Wrapf(err, "failed to click menu: %s", string(out)))

		fmt.Fprintln(logOut, "level3 send")
	}()

	go func() { // level 4
		<-startLvl4
		lvl++
		msgs <- Action{
			Type: "update-item", SeqID: 0,
			Item: Item{Title: "cya..!", Enabled: true},
		}
		msgs <- Action{
			Type: "update-item", SeqID: 2,
			Item: Item{Title: "quit", Enabled: true},
		}

		xdt := exec.Command("xdotool",
			// activate twice to work around render bug...
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"mousemove", thx, thy, "sleep", "1", "click", "1",
			"sleep", "1",
			"mousemove_relative", "--", "0", "-20",
			"click", "1")
		out, err := xdt.CombinedOutput()
		check(errors.Wrapf(err, "failed to click menu: %s", string(out)))

		fmt.Fprintln(logOut, "level4 send")
	}()

	fmt.Fprintln(logOut, "waiting for trayhelper")
	err = th.Wait()
	r.NoError(err)
	r.Equal(4, lvl)

	if _, ok := os.LookupEnv("TRAY_I3"); ok {
		i3exit := exec.Command("i3-msg", "exit")
		i3exit.Stdout = logOut
		i3exit.Stderr = logOut
		err = i3exit.Run()
		t.Log("i3-msg:", err)
		//r.NoError(err, "failed to start the shut down i3")

		t.Log("waiting for i3")
		err = i3.Wait()
	}

	if ffmpeg != nil {
		ffmpeg.Process.Signal(os.Interrupt)
		err := ffmpeg.Wait()
		t.Log("ffmpeg err:", err) // 255? look for "exited normaly"
	}

	if xvfb != nil {
		t.Log("waiting for xvfb")
		err = xvfb.Wait()
		t.Log("xvfb err: err")
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "go routine check failed: %+v\n", err)
		debug.PrintStack()
		os.Exit(1)
	}
}

func halt(cmd *exec.Cmd) {
	fmt.Fprintln(os.Stderr, "halting:", cmd.Path)
	cmd.Process.Kill()
}

// meh - move main to cmd/jsonhelper
type Action struct {
	Type  string `json:"type"`
	Item  Item   `json:"item"`
	Menu  Menu   `json:"menu"`
	SeqID int    `json:"seq_id"`
}
type Item struct {
	Title   string `json:"title"`
	Tooltip string `json:"tooltip"`
	Enabled bool   `json:"enabled"`
	Checked bool   `json:"checked"`
	Hidden  bool   `json:"hidden"`
}
type Menu struct {
	Icon    string `json:"icon"`
	Title   string `json:"title"`
	Tooltip string `json:"tooltip"`
	Items   []Item `json:"items"`
}
