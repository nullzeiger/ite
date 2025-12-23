// Copyright 2025 Ivan Guerreschi.
// BSD-style license.

// Package main implements ITE (Ivan Text Editor), a lightweight Go editor
// built using the modernc.org/tk9.0 GUI toolkit. It provides basic text editing,
// file management, and integrated Go build/run capabilities.
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

// -------------------------------------------------------------------------
// Theme Configuration
// -------------------------------------------------------------------------

// Color palette definitions used throughout the application to maintain
// a consistent visual theme.
const (
	colBlack        = "#000000"
	colExtremeBlack = "#101010"
	colApricotWhite = "#ffffea" // Main background for text areas
	colCoolYellow   = "#eceb91" // Selection background
	colWaterDew     = "#eaffff" // Frame backgrounds
	colSaltWater    = "#d4ffff" // Active button background
	colHighBall     = "#8d8c39" // Scrollbar trough
	colRed          = "#ff0000" // Error/Unsaved status
	colDarkGreen    = "#006400" // Saved status
)

// -------------------------------------------------------------------------
// Application Configuration
// -------------------------------------------------------------------------

const (
	defaultWindowSize    = "1250x600"
	pollInterval         = 100 * time.Millisecond // Frequency for checking build output
	buildChannelBuffer   = 1                      // Buffer size for async command output
	defaultFilePerms     = 0644                   // -rw-r--r--
	defaultFileExtension = ".go"
)

// UI Status messages displayed in the bottom bar or window title.
const (
	statusUntitled = "Untitled - ITE"
	statusNotSaved = "Not saved"
	statusSaved    = "Saved"
	statusBuilding = "Building...\n"
	statusRunning  = "Running...\n"
	statusNoFile   = "No file open. Please save first."
)

// -------------------------------------------------------------------------
// Application State
// -------------------------------------------------------------------------

// Ite represents the main application instance. It holds references to
// UI widgets and manages the application state (current file, build processes).
type Ite struct {
	// Editor components
	editFrame       *TFrameWidget
	editFrame2      *TFrameWidget
	toolbarFrame    *TFrameWidget
	editText        *TextWidget       // Main code editor
	editText2       *TextWidget       // Output console
	editVScrollbar  *TScrollbarWidget // Editor scrollbar
	editVScrollbar2 *TScrollbarWidget // Console scrollbar

	// Status bar components
	statusFrame       *TFrameWidget
	statusLabelCursor *TLabelWidget // Displays Line:Column
	statusLabelFile   *TLabelWidget // Displays Saved/Unsaved status

	// Internal State
	currentFile string      // Absolute path to the currently open file
	buildChan   chan string // Channel to pass async command output to the UI thread
}

// main is the entry point of the application.
func main() {
	NewIte().Run()
}

// Run initializes the window geometry and enters the main Tk event loop.
// This method blocks until the window is closed.
func (i *Ite) Run() {
	WmGeometry(App, defaultWindowSize)
	WmDeiconify(App)
	App.Wait()
}

// NewIte initializes a new instance of the editor.
// It sets up the window title, protocol handlers, widget layout, global styles,
// and starts the background polling loop.
func NewIte() *Ite {
	i := &Ite{
		buildChan: make(chan string, buildChannelBuffer),
	}
	App.WmTitle(statusUntitled)
	// Intercept the close button to prompt for unsaved changes
	WmProtocol(App, "WM_DELETE_WINDOW", i.onQuit)

	i.makeWidgets()
	i.makeLayout()
	i.bindShortcuts()
	i.applyGlobalStyle()

	// Start the polling loop to bridge background goroutines with the UI thread
	TclAfter(pollInterval, i.pollBuildOutput)
	return i
}

// -------------------------------------------------------------------------
// Styling and Layout
// -------------------------------------------------------------------------

// applyGlobalStyle configures the Tcl/Tk theme engine for custom widget appearance.
func (i *Ite) applyGlobalStyle() {
	// Configure Button styles
	StyleConfigure("TButton",
		Background(colWaterDew),
		Foreground(colExtremeBlack),
		Font("GoMono", 11, "bold"))
	StyleMap("TButton", Background, "active", colSaltWater)

	// Configure Scrollbar styles
	StyleConfigure("Vertical.TScrollbar",
		Background(colApricotWhite),
		Troughcolor(colHighBall),
		Borderwidth(1),
		Arrowsize(0))
	StyleMap("TScrollbar", Background, "active", colApricotWhite)

	// Configure Frame and Window background
	StyleConfigure("TFrame", Background(colWaterDew))
	App.Configure(Background(colApricotWhite))
}

// textStyle returns the default configuration options for text widgets.
func textStyle() Opts {
	return Opts{
		Font("GoMono", 13),
		Background(colApricotWhite),
		Foreground(colBlack),
		Insertbackground(colBlack),    // Cursor color
		Selectbackground(colCoolYellow), // Highlight color
		Selectforeground(colBlack),
		Tabs("1c"), // 1 tab width
		Wrap("word"),
		Undo(true), // Enable built-in undo/redo stack
	}
}

// createEditorPanel generates a composite widget containing a text area and
// a vertical scrollbar, properly linked via scroll commands.
func (i *Ite) createEditorPanel() (*TFrameWidget, *TextWidget, *TScrollbarWidget) {
	frame := TFrame()
	text := frame.Text(textStyle(),
		Yscrollcommand(func(event *Event) {
			// This callback updates the scrollbar position when text is scrolled
			// Note: Scrollbar reference is resolved via closure when called
		}))

	scrollbar := frame.TScrollbar(Command(
		func(event *Event) { event.Yview(text) }))

	// Establish the link from Text widget back to Scrollbar
	text.Configure(Yscrollcommand(func(event *Event) {
		event.ScrollSet(scrollbar)
	}))

	return frame, text, scrollbar
}

// makeEditor initializes the main code editing area and the build output console.
func (i *Ite) makeEditor() {
	// Main editor
	i.editFrame, i.editText, i.editVScrollbar = i.createEditorPanel()

	// Output panel
	i.editFrame2, i.editText2, i.editVScrollbar2 = i.createEditorPanel()
}

// makeToolbar creates the top control bar with operation buttons.
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

// makeStatusbar creates the bottom labels for cursor position and file status.
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

// makeWidgets orchestrates the creation of all UI components.
func (i *Ite) makeWidgets() {
	i.makeToolbar()
	i.makeEditor()
	i.makeStatusbar()
}

// makeLayout defines the grid geometry for the main application window.
func (i *Ite) makeLayout() {
	// Toolbar (Row 0, spans entire width)
	Grid(i.toolbarFrame, Row(0), Column(0), Columnspan(2), Sticky(WE))

	// Main Editor Panel (Row 1, Column 0)
	Grid(i.editText, Row(0), Column(0), Sticky(NEWS))
	Grid(i.editVScrollbar, Row(0), Column(1), Sticky(NS))
	GridRowConfigure(i.editFrame, 0, Weight(1))
	GridColumnConfigure(i.editFrame, 0, Weight(1))
	Grid(i.editFrame, Row(1), Column(0), Sticky(NEWS))

	// Output Panel (Row 1, Column 1)
	Grid(i.editText2, Row(0), Column(0), Sticky(NEWS))
	Grid(i.editVScrollbar2, Row(0), Column(1), Sticky(NS))
	GridRowConfigure(i.editFrame2, 0, Weight(1))
	GridColumnConfigure(i.editFrame2, 0, Weight(1))
	Grid(i.editFrame2, Row(1), Column(1), Sticky(NEWS))

	// Status Bar (Row 2, spans entire width)
	Grid(i.statusLabelCursor, Row(0), Column(0), Sticky(WE))
	Grid(i.statusLabelFile, Row(0), Column(1), Sticky(WE))
	GridColumnConfigure(i.statusFrame, 0, Weight(1))
	Grid(i.statusFrame, Row(2), Column(0), Columnspan(2), Sticky(WE))

	// Global Grid Weights (Resizing behavior)
	GridColumnConfigure(App, 0, Weight(1)) // Editor takes 1 part
	GridColumnConfigure(App, 1, Weight(3)) // Output takes 3 parts (seems large, but per design)
	GridRowConfigure(App, 1, Weight(1))    // Content area expands vertically
}

// bindShortcuts maps keyboard shortcuts to application functions.
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
	// Bind cursor movement events to update status bar
	Bind(i.editText, "<ButtonRelease-1>", Command(i.updateCursorPosition))
	Bind(i.editText, "<KeyRelease>", Command(i.updateCursorPosition))
}

// -------------------------------------------------------------------------
// File Operations
// -------------------------------------------------------------------------

// onNew clears the editor to start a new file, checking for unsaved changes first.
func (i *Ite) onNew() {
	if i.promptSaveIfModified() {
		i.editText.Clear()
		i.currentFile = ""
		App.WmTitle(statusUntitled)
		i.editText.SetModified(false)
		i.updateCursorPosition()
	}
}

// onOpen launches a file picker dialog and loads the selected file.
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

// onSave writes the current content to disk. If no file is associated, calls Save As.
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

// onSaveAs launches a file picker to save the content to a new location.
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

// -------------------------------------------------------------------------
// Edit Operations
// -------------------------------------------------------------------------

func (i *Ite) onCut()   { i.editText.Cut() }
func (i *Ite) onCopy()  { i.editText.Copy() }
func (i *Ite) onPaste() { i.editText.Paste() }
func (i *Ite) onUndo()  { i.editText.Undo() }
func (i *Ite) onRedo()  { i.editText.Redo() }

// onQuit attempts to close the application, checking for unsaved changes.
func (i *Ite) onQuit() {
	if i.promptSaveIfModified() {
		Destroy(App)
	}
}

// updateCursorPosition updates the status bar with the current cursor location
// and visual indication of whether the file has been modified (unsaved).
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

// onGoToLine opens a modal dialog allowing the user to jump to a specific line.
func (i *Ite) onGoToLine() {
	dialog := Toplevel()
	dialog.WmTitle("Go to Line")

	// Dialog Layout
	frame := dialog.TFrame()
	Grid(frame, Row(0), Column(0), Padx(10), Pady(10))
	label := frame.TLabel(Txt("Line:Column (e.g. 12.5)"))
	Grid(label, Row(0), Column(0), Sticky(W), Pady(5))
	entry := frame.TEntry(Width(20), Textvariable(""))
	Grid(entry, Row(1), Column(0), Pady(5))
	Focus(entry)

	btnFrame := frame.TFrame()
	Grid(btnFrame, Row(2), Column(0), Pady(10))

	// Navigation Logic
	goToLine := func() {
		input := entry.Textvariable()
		if input == "" {
			return
		}
		// Parse input (supports "line" or "line.col")
		parts := strings.Split(input, ".")
		line := 1
		col := 0
		if len(parts) > 0 {
			fmt.Sscanf(parts[0], "%d", &line)
		}
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &col)
		}
		// Validate bounds
		if line < 1 {
			line = 1
		}
		if col < 0 {
			col = 0
		}
		index := fmt.Sprintf("%d.%d", line, col)

		// Move cursor and scroll
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

	// Dialog shortcuts
	Bind(entry, "<Return>", Command(goToLine))
	Bind(dialog, "<Escape>", Command(func() { Destroy(dialog) }))
}

// -------------------------------------------------------------------------
// Build and Execution Logic
// -------------------------------------------------------------------------

// runCommand executes a Go command asynchronously.
// It sends the resulting output to i.buildChan to be picked up by the UI poller.
func (i *Ite) runCommand(args []string, initialMsg string) {
	// Reset output view
	i.editText2.Configure(State("normal"))
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

		// Format output message
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

		// Non-blocking send to UI channel.
		// If the channel is full, we drain it to ensure the latest message is sent.
		select {
		case i.buildChan <- msg:
		default:
			select {
			case <-i.buildChan: // Drain old message
			default:
			}
			i.buildChan <- msg
		}
	}()
}

// onGoBuild triggers 'go build' on the current project.
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

// onGoRun triggers 'go run' on the current directory.
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

// pollBuildOutput checks the build channel for messages from background goroutines.
// This is necessary because Tk widgets must only be updated from the main thread.
func (i *Ite) pollBuildOutput() {
	select {
	case msg := <-i.buildChan:
		i.editText2.Configure(State("normal"))
		i.editText2.Clear()
		i.editText2.Insert("1.0", msg)
		i.editText2.Configure(State("disabled"))
	default:
		// No messages
	}
	// Schedule next poll
	TclAfter(pollInterval, i.pollBuildOutput)
}

// -------------------------------------------------------------------------
// Helper Functions
// -------------------------------------------------------------------------

// promptSaveIfModified checks if the current file has unsaved changes.
// Returns true if the action can proceed (saved, discarded, or not modified),
// or false if the user cancelled.
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
		return true // Discard changes
	case "cancel":
		return false
	default:
		return false
	}
}

// showError displays a modal error dialog.
func (i *Ite) showError(msg string) {
	MessageBox(Icon("error"), Title("Error"), Msg(msg), Type("ok"))
}
