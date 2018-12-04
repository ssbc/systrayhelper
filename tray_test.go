package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/getlantern/systray"
)

// just so that it compiles and starts
func TestStart(t *testing.T) {

	testf, err := os.Open("test_withshutdown.json")
	if err != nil {
		t.Fatal(err)
	}
	pr, pw := io.Pipe()

	// overwrite main globals
	input = testf
	output = pw

	go func() {
		sc := json.NewDecoder(pr)
		for {
			var a Action
			err := sc.Decode(&a)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(a)
			if a.Type != "clicked" {
				t.Error("did not click?", a.Type)
			}
			if a.Item.Title == "Quit" {
				systray.Quit()
			} else {
				t.Error("clicked wrong item", a.Item)
			}
		}
	}()

	var called = false
	systray.Run(onReady, func() {
		called = true
	})
	if !called {
		t.Error("expected to call onExit")
	}
}

// TODO: integration tests where some1 has to click on the quit-action and the test waits for that..?
