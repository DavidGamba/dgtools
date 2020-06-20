// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package gui...

                ┌──┬──┐
                │  │  │
                ├──┼──┤
                └──┴──┘

*/
package gui

import (
	"fmt"
	"os"
	"strconv"

	"github.com/nsf/termbox-go"
)

type BookmarksGUI struct {
	// Bookmarks map
	Bookmarks map[string]string
	// bookmarks keys list
	Aliases []string
	// Longest alias width
	AliasWidth int
	// currentKey
	currentKey string
	// currentY position
	currentY int
	// Number of lines to offset output
	headerOffset              int
	windowWidth, windowHeight int
	Debug                     bool
}

// New - Returns a BookmarksGUI pointer
func New(bookmarks map[string]string) *BookmarksGUI {
	aliasWidth, aliases := aliasInfo(bookmarks)
	return &BookmarksGUI{
		Bookmarks:    bookmarks,
		Aliases:      aliases,
		AliasWidth:   aliasWidth,
		currentY:     3,
		headerOffset: 3,
	}
}

// aliasInfo - Calculate the lenght of the aliases to align the gui table.
// Get bookmark keys.
func aliasInfo(bookmarks map[string]string) (int, []string) {
	aliases := make([]string, len(bookmarks))
	i := 0
	aliasWidth := 0
	for k := range bookmarks {
		aliases[i] = k
		if len(k) > aliasWidth {
			aliasWidth = len(k)
		}
		i++
	}
	return aliasWidth, aliases
}

func (bgui *BookmarksGUI) Run() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer func() {
		if termbox.IsInit {
			termbox.Close()
		}
	}()
	// ESC means KeyEsc, NOT ModAlt modifier for the next keyboard event
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	alias := bgui.listen()
	termbox.Flush()
	termbox.Close()
	fmt.Println(bgui.Bookmarks[alias])
	return nil
}

// bookmarkLine - Generate a bookmark line of the form:
//
//    n   [alias] fylesystem/location
func (bgui *BookmarksGUI) bookmarkLine(alias string, i int) string {
	return fmt.Sprintf(" %-3d │ %-"+strconv.Itoa(bgui.AliasWidth)+"s │ %s", i, alias, bgui.Bookmarks[alias])
}

// tbPrint - Termbox function to fill a cell with text
func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (bgui *BookmarksGUI) redrawAll() {
	const colorDefault = termbox.ColorDefault
	background := colorDefault
	termbox.Clear(colorDefault, colorDefault)
	width, height := termbox.Size()
	drawBorders(width, height)
	drawRowLine(width, height, 2)
	drawColumnLine(width, height, 6)
	drawColumnLine(width, height, 9+bgui.AliasWidth)
	tbPrint(0, 0, colorDefault, colorDefault, "┌── ")
	tbPrint(4, 0, termbox.ColorRed, termbox.ColorYellow, " Quit: 'q' ")
	tbPrint(15, 0, colorDefault, colorDefault, " ── ")
	tbPrint(19, 0, termbox.ColorBlack, termbox.ColorGreen, " Down: 'j' or '↓' ")
	tbPrint(37, 0, colorDefault, colorDefault, " ── ")
	tbPrint(41, 0, termbox.ColorBlack, termbox.ColorGreen, " Up: 'k' or '↑' ")
	if bgui.Debug {
		tbPrint(0, 1, colorDefault, colorDefault, "│")
		tbPrint(2, 1, termbox.ColorBlue, colorDefault, bgui.currentKey)
	} else {
		tbPrint(0, 1, colorDefault, colorDefault, "│")
	}
	for i, alias := range bgui.Aliases {
		if bgui.currentY == i+bgui.headerOffset {
			background = termbox.AttrReverse
		}
		tbPrint(0, i+bgui.headerOffset, colorDefault, colorDefault, "│")
		tbPrint(1, i+bgui.headerOffset, colorDefault, background, bgui.bookmarkLine(alias, i))
		background = colorDefault
	}
	termbox.Flush()
}

func drawBorders(width, height int) {
	const colorDefault = termbox.ColorDefault
	for x := 0; x < width; x++ {
		switch {
		case x == 0:
			termbox.SetCell(x, 0, '┌', colorDefault, colorDefault)
			termbox.SetCell(x, height-1, '└', colorDefault, colorDefault)
		case x == width-1:
			termbox.SetCell(x, 0, '┐', colorDefault, colorDefault)
			termbox.SetCell(x, height-1, '┘', colorDefault, colorDefault)
		default:
			termbox.SetCell(x, 0, '─', colorDefault, colorDefault)
			termbox.SetCell(x, height-1, '─', colorDefault, colorDefault)
		}
	}
	for y := 0; y < height; y++ {
		switch {
		case y == 0 || y == height-1:
		default:
			termbox.SetCell(0, y, '│', colorDefault, colorDefault)
			termbox.SetCell(width-1, y, '│', colorDefault, colorDefault)
		}
	}
}

func drawRowLine(width, height, y int) {
	const colorDefault = termbox.ColorDefault
	for x := 0; x < width; x++ {
		switch {
		case x == 0:
			termbox.SetCell(x, y, '├', colorDefault, colorDefault)
		case x == width-1:
			termbox.SetCell(x, y, '┤', colorDefault, colorDefault)
		default:
			termbox.SetCell(x, y, '─', colorDefault, colorDefault)
		}
	}
}

func drawColumnLine(width, height, x int) {
	const colorDefault = termbox.ColorDefault
	for y := 2; y < height; y++ {
		switch {
		case y == 2:
			termbox.SetCell(x, y, '┬', colorDefault, colorDefault)
		case y == height-1:
			termbox.SetCell(x, y, '┴', colorDefault, colorDefault)
		default:
			termbox.SetCell(x, y, '│', colorDefault, colorDefault)
		}
	}
}

func (bgui *BookmarksGUI) listen() string {
	lines := len(bgui.Aliases)
	bgui.redrawAll()
mainLoop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainLoop
			case termbox.KeyArrowUp:
				bgui.currentY--
				if bgui.currentY < bgui.headerOffset {
					bgui.currentY = bgui.headerOffset
				}
				bgui.currentKey = fmt.Sprintf("↑: %d", bgui.currentY-bgui.headerOffset)
			case termbox.KeyArrowDown:
				bgui.currentY++
				if bgui.currentY > lines-1+bgui.headerOffset {
					bgui.currentY = lines - 1 + bgui.headerOffset
				}
				bgui.currentKey = fmt.Sprintf("↓: %d", bgui.currentY-bgui.headerOffset)
			case termbox.KeyArrowLeft, termbox.KeyCtrlB:
				bgui.currentKey = "KeyArrowLeft"
			case termbox.KeyArrowRight, termbox.KeyCtrlF:
				bgui.currentKey = "KeyArrowRight"
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				bgui.currentKey = "KeyBackspace"
			case termbox.KeyDelete, termbox.KeyCtrlD:
				bgui.currentKey = "KeyDelete"
			case termbox.KeyTab:
				bgui.currentKey = "KeyTab"
			case termbox.KeySpace:
				bgui.currentKey = "KeySpace"
			case termbox.KeyCtrlK:
				bgui.currentKey = "KeyCtrlK"
			case termbox.KeyEnter:
				bgui.currentKey = "KeyEnter"
				return bgui.Aliases[bgui.currentY-bgui.headerOffset]
			case termbox.KeyHome, termbox.KeyCtrlA:
				bgui.currentKey = "KeyHome"
			case termbox.KeyEnd, termbox.KeyCtrlE:
				bgui.currentKey = "KeyEnd"
			default:
				switch string(ev.Ch) {
				case "q":
					break mainLoop
				case "k":
					bgui.currentY--
					if bgui.currentY < bgui.headerOffset {
						bgui.currentY = bgui.headerOffset
					}
					bgui.currentKey = fmt.Sprintf("k: %d", bgui.currentY-bgui.headerOffset)
				case "j":
					bgui.currentY++
					if bgui.currentY > lines-1+bgui.headerOffset {
						bgui.currentY = lines - 1 + bgui.headerOffset
					}
					bgui.currentKey = fmt.Sprintf("j: %d", bgui.currentY-bgui.headerOffset)
				default:
					bgui.currentKey = string(ev.Ch)
				}
			}
		case termbox.EventMouse:
			button := ""
			switch ev.Key {
			case termbox.MouseLeft:
				button = "MouseLeft"
				if ev.MouseY >= bgui.headerOffset {
					bgui.currentKey = fmt.Sprintf("Mouse event: %d x %d, button: %s", ev.MouseX, ev.MouseY, button)
					bgui.currentY = ev.MouseY
				}
			case termbox.MouseMiddle:
				button = "MouseMiddle"
			case termbox.MouseRight:
				button = "MouseRight"
			case termbox.MouseWheelUp:
				button = "MouseWheelUp"
			case termbox.MouseWheelDown:
				button = "MouseWheelDown"
			case termbox.MouseRelease:
				button = "MouseRelease"
			}
			bgui.currentKey = fmt.Sprintf("Mouse event: %d x %d, button: %s", ev.MouseX, ev.MouseY, button)
		case termbox.EventResize:
		case termbox.EventError:
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", ev.Err)
			break mainLoop
		}
		bgui.redrawAll()
	}
	return ""
}
