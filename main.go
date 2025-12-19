// Copyright 2025 Ivan Guerreschi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "modernc.org/tk9.0"
)

const (
	colBlack        = "#000000"
	colExtremeBlack = "#101010"
	colApricotWhite = "#ffffea"
	colCoolYellow   = "#eceb91"
	colWaterDew     = "#eaffff"
	colSaltWater    = "#d4ffff"
	colHighBall     = "#8d8c39"
	colRed          = "#ff0000"
	colDarkGreen    = "#006400"
)

type Ite struct {
	editFrame         *TFrameWidget
	editFrame2        *TFrameWidget
	toolbarFrame      *TFrameWidget
	editText          *TextWidget
	editText2         *TextWidget
	editVScrollbar    *TScrollbarWidget
	editVScrollbar2   *TScrollbarWidget
	currentFile       string
	statusFrame       *TFrameWidget
	statusLabelCursor *TLabelWidget
	statusLabelFile   *TLabelWidget
}

func main() {
	NewIte().Run()
}

func (i *Ite) Run() {
	WmGeometry(App, "1200x600")
	WmDeiconify(App)
	App.Wait()
}

func NewIte() *Ite {
	i := &Ite{}
	App.WmTitle("Untitled - ITE")
	WmProtocol(App, "WM_DELETE_WINDOW", i.onQuit)
	i.makeWidgets()
	i.makeLayout()
	i.bindShortcuts()

	i.globalStyle()
	return i
}

func (i *Ite) globalStyle() {
	StyleConfigure("TButton", Background(colWaterDew),
		Foreground(colExtremeBlack),
		Font("GoMono", 11, "bold"))
	StyleMap("TButton", Background, "active", colSaltWater)
	StyleConfigure("Vertical.TScrollbar",
		Background(colApricotWhite),
		Troughcolor(colHighBall),
		Borderwidth(1),
		Arrowsize(0))
	StyleConfigure("TFrame", Background(colWaterDew))
	StyleMap("TScrollbar", Background, "active", colApricotWhite)
	App.Configure(Background(colApricotWhite))
}

func textStyle() Opts {
	return Opts{
		Font("GoMono", 13),
		Background(colApricotWhite),
		Foreground(colBlack),
		Insertbackground(colBlack),
		Selectbackground(colCoolYellow),
		Selectforeground(colBlack),
		Tabs("1c"),
		Wrap("word"),
		Undo(true),
	}
}

func statusStyle() Opts {
	return Opts{
		Background(colApricotWhite),
	}
}

func (i *Ite) makeEditor() {
	i.editFrame = TFrame()
	i.editVScrollbar = i.editFrame.TScrollbar()
	i.editText = i.editFrame.Text(
		textStyle(),
		Yscrollcommand(func(event *Event) {
			event.ScrollSet(i.editVScrollbar)
		}))
	i.editVScrollbar = i.editFrame.TScrollbar(Command(
		func(event *Event) { event.Yview(i.editText) }))

	i.editFrame2 = TFrame()
	i.editVScrollbar2 = i.editFrame2.TScrollbar()
	i.editText2 = i.editFrame2.Text(
		textStyle(),
		Yscrollcommand(func(event *Event) {
			event.ScrollSet(i.editVScrollbar2)
		}))
	i.editVScrollbar2 = i.editFrame2.TScrollbar(Command(
		func(event *Event) { event.Yview(i.editText2) }))
}

func (i *Ite) makeToolbar() {
	i.toolbarFrame = TFrame(Relief(RAISED))

	buttons := []struct {
		text string
		cmd  func()
	}{
		{"New", i.onNew},
		{"Open", i.onOpen},
		{"Save", i.onSave},
		{"Cut", i.onCut},
		{"Copy", i.onCopy},
		{"Paste", i.onPaste},
		{"Go to Line", i.onGoToLine},
		{"Go Build", i.onGoBuild},
		{"Go Run", i.onGoRun},
		{"Exit", i.onQuit},
	}

	for col, btn := range buttons {
		b := i.toolbarFrame.TButton(Txt(btn.text), Command(btn.cmd))
		Grid(b, Row(0), Column(col), Sticky(W))
	}
}

func (i *Ite) makeStatusbar() {
	i.statusFrame = TFrame(Relief(SUNKEN))
	i.statusLabelCursor = i.statusFrame.TLabel(Txt("Line:Column 0:0"), Background(colApricotWhite), Font("GoMono", 11))
	i.statusLabelFile = i.statusFrame.TLabel(Txt("Not saved"), Background(colApricotWhite), Font("GoMono", 11))
}

func (i *Ite) makeWidgets() {
	i.makeToolbar()
	i.makeEditor()
	i.makeStatusbar()
}

func (i *Ite) makeLayout() {
	Grid(i.toolbarFrame, Row(0), Column(0), Columnspan(2), Sticky(WE))

	Grid(i.editText, Row(0), Column(0), Sticky(NEWS))
	Grid(i.editVScrollbar, Row(0), Column(1), Sticky(NS))
	GridRowConfigure(i.editFrame, 0, Weight(1))
	GridColumnConfigure(i.editFrame, 0, Weight(1))
	Grid(i.editFrame, Row(1), Column(0), Sticky(NEWS))

	Grid(i.editText2, Row(0), Column(0), Sticky(NEWS))
	Grid(i.editVScrollbar2, Row(0), Column(1), Sticky(NS))
	GridRowConfigure(i.editFrame2, 0, Weight(1))
	GridColumnConfigure(i.editFrame2, 0, Weight(1))
	Grid(i.editFrame2, Row(1), Column(1), Sticky(NEWS))

	Grid(i.statusLabelCursor, Row(0), Column(0), Sticky(WE))
	Grid(i.statusLabelFile, Row(0), Column(1), Sticky(WE))
	GridColumnConfigure(i.statusFrame, 0, Weight(1))
	Grid(i.statusFrame, Row(2), Column(0), Columnspan(2), Sticky(WE))

	GridColumnConfigure(App, 0, Weight(1))
	GridColumnConfigure(App, 1, Weight(3))
	GridRowConfigure(App, 1, Weight(1))
}

func (i *Ite) bindShortcuts() {
	Bind(App, "<Control-n>", Command(i.onNew))
	Bind(App, "<Control-o>", Command(i.onOpen))
	Bind(App, "<Control-s>", Command(i.onSave))
	Bind(App, "<Control-q>", Command(i.onQuit))
	Bind(App, "<Control-g>", Command(i.onGoToLine))
	Bind(i.editText, "<ButtonRelease-1>", Command(i.updateCursorPosition))
	Bind(i.editText, "<KeyRelease>", Command(i.updateCursorPosition))
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

func (i *Ite) updateCursorPosition() {
	pos := i.editText.Index("insert")
	i.statusLabelCursor.Configure(Txt("Line:Column " + pos))

	if i.editText.Modified() {
		i.statusLabelFile.Configure(Foreground(colRed))
		i.statusLabelFile.Configure(Txt("Not saved"))
	} else {
		i.statusLabelFile.Configure(Foreground(colDarkGreen))
		i.statusLabelFile.Configure(Txt("Saved"))
	}
}

func (i *Ite) onGoToLine() {
	dialog := Toplevel()
	dialog.WmTitle("Go to Line")

	frame := dialog.TFrame()
	Grid(frame, Row(0), Column(0), Padx(10), Pady(10))

	label := frame.TLabel(Txt("Line:Column (e.g. 12.5)"))
	Grid(label, Row(0), Column(0), Sticky(W), Pady(5))

	entry := frame.TEntry(Width(20), Textvariable(""))
	Grid(entry, Row(1), Column(0), Pady(5))
	Focus(entry)

	btnFrame := frame.TFrame()
	Grid(btnFrame, Row(2), Column(0), Pady(10))

	okBtn := btnFrame.TButton(Txt("OK"), Command(func() {
		input := entry.Textvariable()

		if input == "" {
			return
		}

		line := 1
		col := 0

		if strings.Contains(input, ".") {
			fmt.Sscanf(input, "%d.%d", &line, &col)
		} else {
			fmt.Sscanf(input, "%d", &line)
		}

		index := fmt.Sprintf("%d.%d", line, col)

		i.editText.MarkSet("insert", index)
		i.editText.See(index)
		i.updateCursorPosition()

		Destroy(dialog)
		Focus(i.editText)
	}))
	Grid(okBtn, Row(0), Column(0), Padx(5))

	cancelBtn := btnFrame.TButton(Txt("Cancel"), Command(func() {
		Destroy(dialog)
	}))
	Grid(cancelBtn, Row(0), Column(1), Padx(5))

	Bind(entry, "<Return>", Command(func() { okBtn.Invoke() }))
}

func (i *Ite) onGoBuild() {
	i.editText2.Clear()
	
	cmd := exec.Command("/usr/local/go/bin/go", "build", "./...")
	output, err := cmd.CombinedOutput()
	
	i.editText2.Clear()
	if err != nil {
		i.editText2.Insert("1.0", fmt.Sprintf("Build failed:\n%s\n", string(output)))
	} else {
		i.editText2.Insert("1.0", "Build successful!\n")
	}
}

func (i *Ite) onGoRun() {
	i.editText2.Clear()
	
	cmd := exec.Command("/usr/local/go/bin/go", "run", "./...")
	output, err := cmd.CombinedOutput()
	
	i.editText2.Clear()
	if err != nil {
		i.editText2.Insert("1.0", fmt.Sprintf("Run failed:\n%s\n", string(output)))
	} else {
		i.editText2.Insert("1.0", fmt.Sprintf("Program output:\n%s", string(output)))
	}
}
