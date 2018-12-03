package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
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
	SeqID uint   `json:"seq_id"`
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
				fmt.Fprintln(os.Stderr, "Unhandled signal:", sig)
			}
		}
	}()
	fmt.Fprintf(os.Stderr, "systrayhelper %s (%s built %s)\n", version, commit, date)

	// We can manipulate the systray in other goroutines
	go func() {
		var items []*systray.MenuItem

		fmt.Fprintln(output, `{"type": "ready"}`)

		//stdinDec := json.NewDecoder(io.TeeReader(input, os.Stderr))
		stdinDec := json.NewDecoder(input)
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
		//fmt.Fprintln(os.Stderr, "menu:", menu)
		icon, err := base64.StdEncoding.DecodeString(menu.Icon)
		if err != nil {
			err = errors.Wrap(err, "failed to decode initial b64 menu icon")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if len(icon) == 0 {
			err = errors.Errorf("menu icon does not contain any bytes")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		systray.SetIcon(icon)
		systray.SetTitle(menu.Title)
		systray.SetTooltip(menu.Tooltip)

		updateItem := func(id uint, item Item) {
			if id >= uint(len(items)) {
				fmt.Fprintf(os.Stderr, "update-item warning!\nSeqID too large - use append-item")
				return
			}
			menuItem := items[id]
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
			menu.Items[id] = item
		}

		updateMenu := func(m Menu) {
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

		appendItem := func(it Item) {
			menuItem := systray.AddMenuItem(it.Title, it.Tooltip)
			if it.Checked {
				menuItem.Check()
			} else {
				menuItem.Uncheck()
			}
			if it.Enabled {
				menuItem.Enable()
			} else {
				menuItem.Disable()
			}
			i := uint(len(items))
			items = append(items, menuItem)

			go func() {
				for range menuItem.ClickedCh {
					err := stdoutEnc.Encode(Action{
						Type:  "clicked",
						Item:  menu.Items[i], // keeps updated title
						SeqID: i,
					})
					fmt.Fprintln(os.Stderr, "clicked err:", err)
				}
			}()
		}

		go func() {
			for {
				var action Action
				if err := stdinDec.Decode(&action); err != nil {
					if err == io.EOF {
						fmt.Fprint(os.Stderr, "trayhelper warning: input decoder loop exited with EOF\n")
						os.Exit(0)
					}
					err = errors.Wrap(err, "failed to decode action")
					fmt.Fprint(os.Stderr, "action loop error:", err)
				}
				fmt.Fprintln(os.Stderr, "got action:", action)
				switch action.Type {
				case "append-item":
					appendItem(action.Item)
					menu.Items = append(menu.Items, action.Item)
				case "update-item":
					updateItem(action.SeqID, action.Item)
				case "update-menu":
					updateMenu(action.Menu)
				case "shutdown":
					if action.SeqID == 999 { // testing magick - testuser could quit faster by clicking
						fmt.Fprintf(os.Stderr, "shutodwn called, still waiting...")
						time.Sleep(time.Second * 5)
						systray.Quit()
					}
				}
			}
		}()

		for _, item := range menu.Items {
			appendItem(item)
		}
	}()
}
