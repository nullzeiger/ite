// Copyright 2025 Ivan Guerreschi.
// BSD-style license.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "modernc.org/tk9.0"
)

// Color scheme
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

// App constants
const (
	defaultWindowSize    = "1200x600"
	pollInterval         = 100 * time.Millisecond
	buildChannelBuffer   = 1
	defaultFilePerms     = 0644
	defaultFileExtension = ".go"
)

// Status messages
const (
	statusUntitled = "Untitled - ITE"
	statusNotSaved = "Not saved"
	statusSaved    = "Saved"
	statusBuilding = "Building...\n"
	statusRunning  = "Running...\n"
	statusNoFile   = "No file open. Please save first."
)

type Ite struct {
	// Editor components
	editFrame       *TFrameWidget
	editFrame2      *TFrameWidget
	toolbarFrame    *TFrameWidget
	editText        *TextWidget
	editText2       *TextWidget
	editVScrollbar  *TScrollbarWidget
	editVScrollbar2 *TScrollbarWidget

	// Status bar
	statusFrame       *TFrameWidget
	statusLabelCursor *TLabelWidget
	statusLabelFile   *TLabelWidget

	// State
	currentFile string
	buildChan   chan string
}

func main() {
	NewIte().Run()
}

func (i *Ite) Run() {
	WmGeometry(App, defaultWindowSize)
	WmDeiconify(App)
	App.Wait()
}

func NewIte() *Ite {
	i := &Ite{
		buildChan: make(chan string, buildChannelBuffer),
	}
	App.WmTitle(statusUntitled)
	WmProtocol(App, "WM_DELETE_WINDOW", i.onQuit)
	i.makeWidgets()
	i.makeLayout()
	i.bindShortcuts()
	i.applyGlobalStyle()
	TclAfter(pollInterval, i.pollBuildOutput)
	return i
}

// applyGlobalStyle configures the global theme
func (i *Ite) applyGlobalStyle() {
	StyleConfigure("TButton",
		Background(colWaterDew),
		Foreground(colExtremeBlack),
		Font("GoMono", 11, "bold"))
	StyleMap("TButton", Background, "active", colSaltWater)
	StyleConfigure("Vertical.TScrollbar",
		Background(colApricotWhite),
		Troughcolor(colHighBall),
		Borderwidth(1),
		Arrowsize(0))
	StyleMap("TScrollbar", Background, "active", colApricotWhite)
	StyleConfigure("TFrame", Background(colWaterDew))
	App.Configure(Background(colApricotWhite))
}

// textStyle returns common text widget options
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

// createEditorPanel creates a text editor with scrollbar
func (i *Ite) createEditorPanel() (*TFrameWidget, *TextWidget, *TScrollbarWidget) {
	frame := TFrame()
	text := frame.Text(textStyle(),
		Yscrollcommand(func(event *Event) {
			// Scrollbar will be set after creation
		}))
	scrollbar := frame.TScrollbar(Command(
		func(event *Event) { event.Yview(text) }))
	// Update the scroll command now that scrollbar exists
	text.Configure(Yscrollcommand(func(event *Event) {
		event.ScrollSet(scrollbar)
	}))
	return frame, text, scrollbar
}

func (i *Ite) makeEditor() {
	// Main editor
	i.editFrame, i.editText, i.editVScrollbar = i.createEditorPanel()
	// Output panel (read-only by default)
	i.editFrame2, i.editText2, i.editVScrollbar2 = i.createEditorPanel()
	i.editText2.Configure(State("disabled")) // Make output read-only
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
		{"Save As", i.onSaveAs},
		{"Cut", i.onCut},
		{"Copy", i.onCopy},
		{"Paste", i.onPaste},
		{"Undo", i.onUndo},
		{"Redo", i.onRedo},
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
	i.statusLabelCursor = i.statusFrame.TLabel(
		Txt("Line:Column 0:0"),
		Background(colApricotWhite),
		Font("GoMono", 11))
	i.statusLabelFile = i.statusFrame.TLabel(
		Txt(statusNotSaved),
		Background(colApricotWhite),
		Font("GoMono", 11))
}

func (i *Ite) makeWidgets() {
	i.makeToolbar()
	i.makeEditor()
	i.makeStatusbar()
}

func (i *Ite) makeLayout() {
	// Toolbar
	Grid(i.toolbarFrame, Row(0), Column(0), Columnspan(2), Sticky(WE))
	// Main editor panel
	Grid(i.editText, Row(0), Column(0), Sticky(NEWS))
	Grid(i.editVScrollbar, Row(0), Column(1), Sticky(NS))
	GridRowConfigure(i.editFrame, 0, Weight(1))
	GridColumnConfigure(i.editFrame, 0, Weight(1))
	Grid(i.editFrame, Row(1), Column(0), Sticky(NEWS))
	// Output panel
	Grid(i.editText2, Row(0), Column(0), Sticky(NEWS))
	Grid(i.editVScrollbar2, Row(0), Column(1), Sticky(NS))
	GridRowConfigure(i.editFrame2, 0, Weight(1))
	GridColumnConfigure(i.editFrame2, 0, Weight(1))
	Grid(i.editFrame2, Row(1), Column(1), Sticky(NEWS))
	// Status bar
	Grid(i.statusLabelCursor, Row(0), Column(0), Sticky(WE))
	Grid(i.statusLabelFile, Row(0), Column(1), Sticky(WE))
	GridColumnConfigure(i.statusFrame, 0, Weight(1))
	Grid(i.statusFrame, Row(2), Column(0), Columnspan(2), Sticky(WE))
	// Main window layout
	GridColumnConfigure(App, 0, Weight(1))
	GridColumnConfigure(App, 1, Weight(3))
	GridRowConfigure(App, 1, Weight(1))
}

func (i *Ite) bindShortcuts() {
	shortcuts := map[string]func(){
		"<Control-n>":       i.onNew,
		"<Control-o>":       i.onOpen,
		"<Control-s>":       i.onSave,
		"<Control-Shift-s>": i.onSaveAs,
		"<Control-q>":       i.onQuit,
		"<Control-b>":       i.onGoBuild,
		"<Control-r>":       i.onGoRun,
		"<Control-g>":       i.onGoToLine,
		"<Control-z>":       i.onUndo,
		"<Control-y>":       i.onRedo,
	}
	for key, cmd := range shortcuts {
		Bind(App, key, Command(cmd))
	}
	Bind(i.editText, "<ButtonRelease-1>", Command(i.updateCursorPosition))
	Bind(i.editText, "<KeyRelease>", Command(i.updateCursorPosition))
}

func (i *Ite) onNew() {
	if i.promptSaveIfModified() {
		i.editText.Clear()
		i.currentFile = ""
		App.WmTitle(statusUntitled)
		i.editText.SetModified(false)
		i.updateCursorPosition()
	}
}

func (i *Ite) onOpen() {
	if !i.promptSaveIfModified() {
		return
	}
	paths := GetOpenFile(Title("Open"), Initialdir("."), Filetypes([]FileType{
		{TypeName: "Go Files", Extensions: []string{"*.go"}, MacType: ""},
		{TypeName: "All Files", Extensions: []string{"*"}, MacType: ""},
	}))
	if len(paths) == 0 {
		return
	}
	path := paths[0]
	data, err := os.ReadFile(path)
	if err != nil {
		i.showError("Error opening file: " + err.Error())
		return
	}
	i.editText.Clear()
	i.editText.Insert("1.0", string(data))
	i.currentFile = path
	App.WmTitle(fmt.Sprintf("%s - ITE", filepath.Base(i.currentFile)))
	i.editText.SetModified(false)
	i.updateCursorPosition()
}

func (i *Ite) onSave() {
	if i.currentFile == "" {
		i.onSaveAs()
		return
	}
	content := i.editText.Text()
	if err := os.WriteFile(i.currentFile, []byte(content), defaultFilePerms); err != nil {
		i.showError("Error saving file: " + err.Error())
		return
	}
	App.WmTitle(fmt.Sprintf("%s - ITE", filepath.Base(i.currentFile)))
	i.editText.SetModified(false)
	i.updateCursorPosition()
}

func (i *Ite) onSaveAs() {
	path := GetSaveFile(Title("Save as..."), Initialdir("."), Filetypes([]FileType{
		{TypeName: "Go Files", Extensions: []string{"*.go"}, MacType: ""},
		{TypeName: "All Files", Extensions: []string{"*"}, MacType: ""},
	}))
	if path == "" {
		return
	}
	if filepath.Ext(path) == "" {
		path += defaultFileExtension
	}
	i.currentFile = path
	i.onSave()
}

func (i *Ite) onCut()   { i.editText.Cut() }
func (i *Ite) onCopy()  { i.editText.Copy() }
func (i *Ite) onPaste() { i.editText.Paste() }
func (i *Ite) onUndo()  { i.editText.Undo() }
func (i *Ite) onRedo()  { i.editText.Redo() }

func (i *Ite) onQuit() {
	if i.promptSaveIfModified() {
		Destroy(App)
	}
}

func (i *Ite) updateCursorPosition() {
	pos := i.editText.Index("insert")
	i.statusLabelCursor.Configure(Txt("Line:Column " + pos))
	if i.editText.Modified() {
		i.statusLabelFile.Configure(
			Foreground(colRed),
			Txt(statusNotSaved))
	} else {
		i.statusLabelFile.Configure(
			Foreground(colDarkGreen),
			Txt(statusSaved))
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
	goToLine := func() {
		input := entry.Textvariable()
		if input == "" {
			return
		}
		parts := strings.Split(input, ".")
		line := 1
		col := 0
		if len(parts) > 0 {
			fmt.Sscanf(parts[0], "%d", &line)
		}
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &col)
		}
		if line < 1 {
			line = 1
		}
		if col < 0 {
			col = 0
		}
		index := fmt.Sprintf("%d.%d", line, col)
		i.editText.MarkSet("insert", index)
		i.editText.See(index)
		i.updateCursorPosition()
		Destroy(dialog)
		Focus(i.editText)
	}
	okBtn := btnFrame.TButton(Txt("OK"), Command(goToLine))
	Grid(okBtn, Row(0), Column(0), Padx(5))
	cancelBtn := btnFrame.TButton(Txt("Cancel"), Command(func() {
		Destroy(dialog)
	}))
	Grid(cancelBtn, Row(0), Column(1), Padx(5))
	Bind(entry, "<Return>", Command(goToLine))
	Bind(dialog, "<Escape>", Command(func() { Destroy(dialog) }))
}

// runCommand executes a Go command and sends output to the build channel
func (i *Ite) runCommand(args []string, initialMsg string) {
	i.editText2.Configure(State("normal")) // Temporarily enable for insertion
	i.editText2.Clear()
	i.editText2.Insert("1.0", initialMsg)
	i.editText2.Configure(State("disabled"))
	go func() {
		cmd := exec.Command("go", args...)
		if i.currentFile != "" {
			cmd.Dir = filepath.Dir(i.currentFile)
		}
		output, err := cmd.CombinedOutput()
		var msg string
		if err != nil {
			if len(output) == 0 {
				msg = fmt.Sprintf("%s failed: %v\n", args[0], err)
			} else {
				msg = fmt.Sprintf("%s failed:\n%s", strings.Title(args[0]), string(output))
			}
		} else if len(output) == 0 {
			msg = fmt.Sprintf("%s successful\n", strings.Title(args[0]))
		} else {
			msg = fmt.Sprintf("%s output:\n%s", strings.Title(args[0]), string(output))
		}
		// Non-blocking send with replacement if full
		select {
		case i.buildChan <- msg:
		default:
			select {
			case <-i.buildChan:
			default:
			}
			i.buildChan <- msg
		}
	}()
}

func (i *Ite) onGoBuild() {
	if i.currentFile == "" {
		i.showError(statusNoFile)
		return
	}
	if i.editText.Modified() {
		i.onSave()
	}
	i.runCommand([]string{"build", "./..."}, statusBuilding)
}

func (i *Ite) onGoRun() {
	if i.currentFile == "" {
		i.showError(statusNoFile)
		return
	}
	if i.editText.Modified() {
		i.onSave()
	}
	i.runCommand([]string{"run", "."}, statusRunning)
}

func (i *Ite) pollBuildOutput() {
	select {
	case msg := <-i.buildChan:
		i.editText2.Configure(State("normal"))
		i.editText2.Clear()
		i.editText2.Insert("1.0", msg)
		i.editText2.Configure(State("disabled"))
	default:
	}
	TclAfter(pollInterval, i.pollBuildOutput)
}

// promptSaveIfModified checks if modified and prompts to save
func (i *Ite) promptSaveIfModified() bool {
	if !i.editText.Modified() {
		return true
	}
	resp := MessageBox(Icon("question"), Title("Unsaved Changes"), Msg("Save changes?"), Detail("Your changes will be lost if you don't save them."), Type("yesnocancel"))
	switch resp {
	case "yes":
		i.onSave()
		return true
	case "no":
		return true
	case "cancel":
		return false
	default:
		return false
	}
}

// showError displays an error message in a dialog
func (i *Ite) showError(msg string) {
	MessageBox(Icon("error"), Title("Error"), Msg(msg), Type("ok"))
}
