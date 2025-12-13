// Copyright 2025 Ivan Guerreschi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	. "modernc.org/tk9.0"
)

const (
	black = "#000000"
	extremeblack = "#101010"
	apricotwhite  = "#ffffea"
	coolyellow = "#eceb91"
	waterdew  = "#eaffff"
	saltwater = "#d4ffff"
	highball = "#8d8c39"
)

type Ite struct {
	toolbarFrame    *TFrameWidget
	newToolButton   *TButtonWidget
	openToolButton  *TButtonWidget
	saveToolButton  *TButtonWidget
	cutToolButton   *TButtonWidget
	copyToolButton  *TButtonWidget
	pasteToolButton *TButtonWidget
	exitToolButton  *TButtonWidget
	editFrame       *TFrameWidget
	editText        *TextWidget
	editVScrollbar  *TScrollbarWidget
	currentFile     string
}

func main() {
	NewIte().Run()
}

func (i *Ite) Run() {
	WmGeometry(App, "860x580")
	App.Center()
	WmDeiconify(App)
	App.Wait()
}

func NewIte() *Ite {
	i := &Ite{}
	i.globalStyle()
	i.makeWidgets()
	i.makeLayout()
	return i
}

func (i *Ite) globalStyle() {
	StyleConfigure("TButton", Background(waterdew),
		Foreground(extremeblack),
		Font("GoMono", 11, "bold"))
	StyleMap("TButton", Background, "active", saltwater)
	StyleConfigure("Vertical.TScrollbar",
		Background(apricotwhite),
		Troughcolor(highball),
		Borderwidth(1),
		Arrowsize(0))
	StyleConfigure("TFrame", Background(waterdew))
	StyleMap("TScrollbar", Background, "active", apricotwhite)
	App.Configure(Background(apricotwhite))
}

func textStyle() Opts {
	return Opts{
		Font("GoMono", 13),
		Background(apricotwhite),
		Foreground(black),
		Insertbackground(black),
		Selectbackground(coolyellow),
		Selectforeground(black),
		Tabs("1c"),
	}
}

func (i *Ite) makeEditor() {
	i.editFrame = TFrame()
	i.editText = i.editFrame.Text(
		textStyle(),
		Yscrollcommand(func(event *Event) {
		event.ScrollSet(i.editVScrollbar)
		}))
	i.editVScrollbar = i.editFrame.TScrollbar(Command(
		func(event *Event) { event.Yview(i.editText) }))
}

func (i *Ite) makeToolbar() {
	i.toolbarFrame = TFrame(Relief(RAISED))
	i.newToolButton = i.toolbarFrame.TButton(Txt("New"), Command(i.onNew))
	i.openToolButton = i.toolbarFrame.TButton(Txt("Open"), Command(i.onOpen))
	i.saveToolButton = i.toolbarFrame.TButton(Txt("Save"), Command(i.onSave))
	i.cutToolButton = i.toolbarFrame.TButton(Txt("Cut"), Command(i.onCut))
	i.copyToolButton = i.toolbarFrame.TButton(Txt("Copy"), Command(i.onCopy))
	i.pasteToolButton = i.toolbarFrame.TButton(Txt("Paste"), Command(i.onPaste))
	i.exitToolButton = i.toolbarFrame.TButton(Txt("Exit"), Command(i.onQuit))
}

func (i *Ite) layoutEditor() {
	Grid(i.editText, Row(0), Column(0), Sticky(NEWS))
	Grid(i.editVScrollbar, Row(0), Column(1), Sticky(NS))
	GridRowConfigure(i.editFrame, 0, Weight(1))
	GridColumnConfigure(i.editFrame, 0, Weight(1))
}

func (i *Ite) layoutToolbar() {
	opts := Opts{Sticky(W)}
	Grid(i.newToolButton, Row(0), Column(1), opts)
	Grid(i.openToolButton, Row(0), Column(2), opts)
	Grid(i.saveToolButton, Row(0), Column(3), opts)
	Grid(i.cutToolButton, Row(0), Column(4), opts)
	Grid(i.copyToolButton, Row(0), Column(5), opts)
	Grid(i.pasteToolButton, Row(0), Column(6), opts)
	Grid(i.exitToolButton, Row(0), Column(7), opts)
}

func (i *Ite) makeWidgets() {
	i.makeToolbar()
	i.makeEditor()
}

func (i *Ite) makeLayout() {
	i.layoutToolbar()
	Grid(i.toolbarFrame, Row(0), Column(0), Sticky(WE))
	i.layoutEditor()
	Grid(i.editFrame, Row(1), Column(0), Sticky(NEWS))
	GridColumnConfigure(App, 0, Weight(1))
	GridRowConfigure(App, 1, Weight(1))
}

func (i *Ite) onNew() {
	i.editText.Clear()
	i.currentFile = ""
	App.WmTitle("Untitled - ITE")
	i.editText.SetModified(false)
}

func (i *Ite) onOpen() {
	paths := GetOpenFile(Title("Open"), Initialdir("."))
	if len(paths) == 0 {
		return
	}
	path := paths[0]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	i.editText.Clear()
	i.editText.Insert("1.0", string(data))
	i.currentFile = path
	App.WmTitle(fmt.Sprintf("%s - ITE", filepath.Base(i.currentFile)))
	i.editText.SetModified(false)
}

func (i *Ite) onSave() {
	if i.currentFile == "" {
		path := GetSaveFile(Title("Save as..."), Initialdir("."))
		if path == "" {
			return
		}
		i.currentFile = path
	}
	content := i.editText.Text()
	if err := os.WriteFile(i.currentFile, []byte(content), 0644); err != nil {
		fmt.Println("Error saving file:", err)
		return
	}
	App.WmTitle(fmt.Sprintf("%s - ITE", filepath.Base(i.currentFile)))
	i.editText.SetModified(false)
}

func (i *Ite) onCut() {
	i.editText.Cut()
}

func (i *Ite) onCopy() {
	i.editText.Copy()
}

func (i *Ite) onPaste() {
	i.editText.Paste()
}

func (i *Ite) onQuit() { Destroy(App) }