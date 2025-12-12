// Copyright 2025 Ivan Guerreschi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import . "modernc.org/tk9.0"

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
	i.makeWidgets()
	i.makeLayout()
	return i
}

func (i *Ite) makeEditor() {
	i.editFrame = TFrame()
	i.editText = i.editFrame.Text(
			Yscrollcommand(func(event *Event) {
			event.ScrollSet(i.editVScrollbar)
		}))
	i.editVScrollbar = i.editFrame.TScrollbar(Command(
		func(event *Event) { event.Yview(i.editText) }))
}

func (i *Ite) makeToolbar() {
	i.toolbarFrame = TFrame(Relief(RAISED))
	i.newToolButton = i.toolbarFrame.TButton(Txt("New"))
	i.openToolButton = i.toolbarFrame.TButton(Txt("Open"))
	i.saveToolButton = i.toolbarFrame.TButton(Txt("Save"))
	i.cutToolButton = i.toolbarFrame.TButton(Txt("Cut"))
	i.copyToolButton = i.toolbarFrame.TButton(Txt("Copy"))
	i.pasteToolButton = i.toolbarFrame.TButton(Txt("Paste"))
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

func (i *Ite) onQuit() { Destroy(App) }