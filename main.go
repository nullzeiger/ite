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