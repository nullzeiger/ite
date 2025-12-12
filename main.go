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
	ite := &Ite{}
	return ite
}

func (i *Ite) makeToolbar() {
	i.toolbarFrame = TFrame(Relief(RAISED))
	i.newToolButton = i.toolbarFrame.TButton(Txt("New"))
	i.openToolButton = i.toolbarFrame.TButton(Txt("Open"))
	i.saveToolButton = i.toolbarFrame.TButton(Txt("Save"))
	i.cutToolButton = i.toolbarFrame.TButton(Txt("Cut"))
	i.copyToolButton = i.toolbarFrame.TButton(Txt("Copy"))
	i.pasteToolButton = i.toolbarFrame.TButton(Txt("Paste"))
	i.exitToolButton = i.toolbarFrame.TButton(Txt("Exit"), Command(func() { Destroy(App) }))
}
