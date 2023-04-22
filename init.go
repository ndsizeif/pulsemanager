// /////////////////////////////////////////////////////////////////////////////
// INITIALIZATION OF PROGRAM
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	tea "github.com/charmbracelet/bubbletea" // main cli application library
	"os"                                     // inferface with operating system
	"os/exec"                                // run external system commands
	"strings"                                // manipulate strings
)

// check for existence of binary on system
func haveProgram(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// check for a running pulseaudio server on system
func havePulseServer(cmd string) bool {
	_, err := exec.Command("pulseaudio", cmd).Output()
	return err == nil
}

// check for a tty terminal by parsing env variable
func isConsole() bool {
	if len(strings.TrimSpace(os.Getenv("DISPLAY"))) == 0 {
		return true
	}
	return false
}
func haveNoColor() bool { // TODO fix bar color adherence to no color
	color := os.Getenv("NO_COLOR")
	if color != "" {
		return true
	}
	return false
}

// unset icon variables or use ascii characters only
func disableSymbols() {
	muted_icon = ""
	unmuted_icon = ""
	sus_icon = ""
	idle_icon = ""
	mic_icon = ""
	pref_icon = ">>> "
	suff_icon = " <<<"
}

// build and return the intial model that will be passed to tea.NewProgram
func setupModel() model {
	w := (setWidth / 4) * 3                          // define value that will truncate strings/bar
	d, dc := buildDevices(buildPulse())              // create PulseDevice from json data
	c := dc.total - dc.cards                         // exclude cards from count
	formatProgressBars(d, colorBars(deviceColor), w) // popluate/style progress bars for devices
	keySetup := initKeymapping()                     // returns address of keymap settings
	pageSetup := initPager(setItems, c)              // returns paginator settings
	borderSetup := initBorder(setBorder)
	return model{
		Device:      d,               // pulseaudio data and progress model
		Count:       dc,              // number of each device type
		Keys:        *keySetup,       // program key bindings
		Paginator:   pageSetup,       // paging model
		Help:        initHelp(),      // help model
		Fullscreen:  setAltscreen,    // program begins in fullscreen
		ShowMessage: !setNoMessages,  // program begins with messages on
		ChannelMode: -1,              // control all channels of device (-1=all)
		VolumeLimit: setMaxVolume,    // set the maximum volume of devices
		StringLen:   w,               // calculate max string/bar length by setWidth
		Border:      borderSetup,     // pass border type from config
		Text:        initText(),      // pass generic lipgloss style for text
		Selected:    initSelection(), // initalize values to -1/empty string
		Cursor:      initCursor(),    // pass initial cursor values, if any
		Display:     initDisplay(),   // pass display attributes
	}
}

// first bubbletea function called, returns optional initial command
func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd             // slice holds multiple commands
	cmds = append(cmds, tickCmd()) // start timer for update interval
	m.Message = errorLoad
	return tea.Batch(cmds...) // use tea.Batch for multiple cmds
}
