package main

import (
	"github.com/jroimartin/gocui"
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("output", 0, 0, maxX-1, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Messages"
		v.Wrap = true
		v.Autoscroll = true
		outputView = v
	}

	if v, err := g.SetView("input", 0, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Enter your message (Press Enter to send, Ctrl+C to exit)"
		v.Editable = true
		v.Wrap = true
		inputView = v
		if e, err := g.SetCurrentView("input"); e == nil || err != nil {
			return err
		}
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
