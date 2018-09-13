package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/pkg/errors"
)

var (
	version = "v0.0.5-snapshot"
	commit  = "unset"
	date    = "unset"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: systrayhelper\n")
		fmt.Fprintf(os.Stderr, "\tsend&receive json over stdio to create menu items, etc.\n\n")
		fmt.Fprintf(os.Stderr, "Version %s (%s built %s)\n", version, commit, date)
		os.Exit(0)
	}
	// Should be called at the very beginning of main().
	systray.Run(onReady, onExit)
}

func onExit() {
	os.Exit(0)
}

// Item represents an item in the menu
type Item struct {
	Title   string `json:"title"`
	Tooltip string `json:"tooltip"`
	Enabled bool   `json:"enabled"`
	Checked bool   `json:"checked"`
}

// Menu has an icon, title and list of items
type Menu struct {
	Icon    string `json:"icon"`
	Title   string `json:"title"`
	Tooltip string `json:"tooltip"`
	Items   []Item `json:"items"`
}

// Action for an item?..
type Action struct {
	Type  string `json:"type"`
	Item  Item   `json:"item"`
	Menu  Menu   `json:"menu"`
	SeqID int    `json:"seq_id"`
}

// test stubbing
var (
	input  io.Reader = os.Stdin
	output io.Writer = os.Stdout
)

func onReady() {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range signalChannel {
			switch sig {
			case os.Interrupt, syscall.SIGTERM:
				//handle SIGINT, SIGTERM
				fmt.Fprintf(os.Stderr, "%s: exiting\n", os.Args[0])
				systray.Quit()
			default:
				fmt.Println("Unhandled signal:", sig)
			}
		}
	}()
	fmt.Fprintf(os.Stderr, "systrayhelper %s (%s built %s)\n", version, commit, date)
	// We can manipulate the systray in other goroutines
	go func() {
		var items []*systray.MenuItem

		fmt.Println(`{"type": "ready"}`)

		stdinDec := json.NewDecoder(input) // debug: io.TeeReader(input, os.Stderr))
		stdoutEnc := json.NewEncoder(output)

		var menu Menu
		err := stdinDec.Decode(&menu)
		if err != nil {
			if err == io.EOF {
				systray.Quit()
				return
			}
			err = errors.Wrap(err, "failed to decode initial menu config")
			fmt.Fprintln(os.Stderr, err)
		}
		// fmt.Println("menu", menu)
		icon, err := base64.StdEncoding.DecodeString(menu.Icon)
		if err != nil {
			err = errors.Wrap(err, "failed to decode initial b64 menu icon")
			fmt.Fprintln(os.Stderr, err)
		}
		systray.SetIcon(icon)
		systray.SetTitle(menu.Title)
		systray.SetTooltip(menu.Tooltip)

		updateItem := func(a Action) {
			if a.SeqID >= len(items) {
				// todo: extend
				fmt.Fprintf(os.Stderr, "update-item warning!\nSeqID too large. has:%d want:%d\n", len(items), a.SeqID)
				return
			}
			item := a.Item
			menuItem := items[a.SeqID]
			if item.Checked {
				menuItem.Check()
			} else {
				menuItem.Uncheck()
			}
			if item.Enabled {
				menuItem.Enable()
			} else {
				menuItem.Disable()
			}
			menuItem.SetTitle(item.Title)
			menuItem.SetTooltip(item.Tooltip)
		}
		updateMenu := func(action Action) {
			m := action.Menu
			if menu.Title != m.Title {
				menu.Title = m.Title
				systray.SetTitle(menu.Title)
			}
			if menu.Icon != m.Icon {
				menu.Icon = m.Icon
				icon, err := base64.StdEncoding.DecodeString(menu.Icon)
				if err != nil {
					err = errors.Wrap(err, "failed to decode b64 menu icon string")
					fmt.Fprintln(os.Stderr, err)
				}
				systray.SetIcon(icon)
			}
			if menu.Tooltip != m.Tooltip {
				menu.Tooltip = m.Tooltip
				systray.SetTooltip(menu.Tooltip)
			}
		}

		update := func(action Action) {
			switch action.Type {
			case "update-item":
				updateItem(action)
			case "update-menu":
				updateMenu(action)
			case "update-item-and-menu":
				updateItem(action)
				updateMenu(action)
			case "shutdown":
				if action.SeqID == 999 { // testing magick - testuser could quit faster by clicking
					fmt.Fprintf(os.Stderr, "shutodwn called, still waiting...")
					time.Sleep(time.Second * 5)
					systray.Quit()
				}
			}
		}

		go func() {
			for {
				var action Action
				if err := stdinDec.Decode(&action); err != nil {
					if err == io.EOF {
						fmt.Fprint(os.Stderr, "trayhelper warning: input decoder loop exited with EOF\n")
						break
					}
					err = errors.Wrap(err, "failed to decode action")
					fmt.Fprint(os.Stderr, "action loop error:", err)
				}
				update(action)
			}
		}()

		for i := 0; i < len(menu.Items); i++ {
			item := menu.Items[i]
			menuItem := systray.AddMenuItem(item.Title, item.Tooltip)
			if item.Checked {
				menuItem.Check()
			} else {
				menuItem.Uncheck()
			}
			if item.Enabled {
				menuItem.Enable()
			} else {
				menuItem.Disable()
			}
			items = append(items, menuItem)
		}

		// {"type": "update-item", "item": {"Title":"aa3","Tooltip":"bb","Enabled":true,"Checked":true}, "seqID": 0}
		for {
			/* this builds a dynamic chan select
			select {
				// for item := range items
				case chosen <- item[i]
			}
			*/
			cases := make([]reflect.SelectCase, len(items))
			for i, ch := range items {
				cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.ClickedCh)}
			}

			remaining := len(cases)
			for remaining > 0 {
				chosen, _, ok := reflect.Select(cases)
				if !ok {
					// The chosen channel has been closed, so zero out the channel to disable the case
					cases[chosen].Chan = reflect.ValueOf(nil)
					remaining--
					continue
				}
				// menuItem := items[chosen]
				err := stdoutEnc.Encode(Action{
					Type:  "clicked",
					Item:  menu.Items[chosen],
					SeqID: chosen,
				})
				if err != nil {
					err = errors.Wrap(err, "failed to encode clicked action")
					fmt.Fprintln(os.Stderr, err)
				}
			}
		}
	}()
}
