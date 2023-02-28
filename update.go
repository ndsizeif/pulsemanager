// /////////////////////////////////////////////////////////////////////////////
// BUBBLETEA UPDATE
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	"fmt"                                    // format and print text
	"github.com/charmbracelet/bubbles/key"   // define application key map
	tea "github.com/charmbracelet/bubbletea" // main cli application library
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd         // create array of tea.Cmds
	var cmd tea.Cmd            // create a tea.Cmd variable
	displayInitMessage(&m)     // show model initialization message
	switch msg := msg.(type) { // parse messages by type
	/////////////////////////////////////////////// INTERVAL REFRESH
	case TickMsg:
		cmd = updateDevices(m)
		cmds = append(cmds, cmd)
		cmds = append(cmds, tickCmd())
		return m, tea.Batch(cmds...)
	case RefreshMsg:
		m.Device = msg.device
		m.Count = msg.count
		refreshPosition(&m)
	////////////////// WINDOW RESIZE ///////////////
	case tea.WindowSizeMsg:
		resizeProgram(&m, msg)
		return m, validateTerminalSize(&m, msg)
	////////////////// KEYSTROKES //////////////////
	case tea.KeyMsg:
		switch {
		//////////////// NAVIGATION //////////////////
		case key.Matches(msg, m.Keys.Up):
			if m.Cursor.pos <= 0 { // prevent model from updating on extraneous up
				return m, nil
			}
			cursorUp(&m, msg)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Down):
			if m.Cursor.pos >= len(m.Device)-1 { // no updating on extraneous down
				return m, nil
			}
			cursorDown(&m, msg)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.GoToStart):
			cursorFirst(&m)
		case key.Matches(msg, m.Keys.GoToEnd):
			cursorLast(&m)
		case key.Matches(msg, m.Keys.PrevPage):
			cursorPage(&m, false)
		case key.Matches(msg, m.Keys.NextPage):
			cursorPage(&m, true)
		case key.Matches(msg, m.Keys.Escape):
			resetOptions(&m)
			return m, updateDevices(m)
		//////////////// COMMANDS ////////////////////
		case key.Matches(msg, m.Keys.PerformAction):
			performAction(&m)
		case key.Matches(msg, m.Keys.ChangeDisplay):
			changeDisplayLevel(&m)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.ChangeChannel):
			changeChannel(&m)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.SelectDevice):
			selectDevice(&m)
		case key.Matches(msg, m.Keys.KillStream):
			killStreamOrOutput(&m)
		case key.Matches(msg, m.Keys.UnloadLoopback):
			unloadLoopback(&m)
		case key.Matches(msg, m.Keys.LatencyUp):
			changeLatency(&m, true)
		case key.Matches(msg, m.Keys.LatencyDown):
			changeLatency(&m, false)
		case key.Matches(msg, m.Keys.Refresh):
		case key.Matches(msg, m.Keys.Fullscreen):
			return toggleFullscreen(&m)
		case key.Matches(msg, m.Keys.ShowMessage):
			showMessages(&m)
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.Keys.ShowFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll
		// case key.Matches(msg, m.Keys.Demo):
			// displayProgramMessage(&m)
		//////////////// VOLUME //////////////////////
		case key.Matches(msg, m.Keys.Mute):
			toggleDeviceMute(&m)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.VolumeUp):
			changeDeviceVolume(&m, formatDeviceVolume(&m, true))
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.VolumeDown):
			changeDeviceVolume(&m, formatDeviceVolume(&m, false))
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume10):
			normalizeDeviceVolume(&m, 10)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume20):
			normalizeDeviceVolume(&m, 20)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume30):
			normalizeDeviceVolume(&m, 30)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume40):
			normalizeDeviceVolume(&m, 40)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume50):
			normalizeDeviceVolume(&m, 50)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume60):
			normalizeDeviceVolume(&m, 60)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume70):
			normalizeDeviceVolume(&m, 70)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume80):
			normalizeDeviceVolume(&m, 80)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume90):
			normalizeDeviceVolume(&m, 90)
			return m, updateDevices(m)
		case key.Matches(msg, m.Keys.Volume100):
			normalizeDeviceVolume(&m, 100)
			return m, updateDevices(m)
		}
	}
	return m, cmd
}

// helper changes cursor to the first device on page when jumping by page
func setCursor(m *model) int {
	cursor := (m.Paginator.Page * m.Paginator.PerPage)
	return cursor
}

// helper sets channel mode to -1/default
func resetChannelMode(m *model) {
	m.ChannelMode = -1
}
func resetSelected(m *model) {
	m.Selected.devicetype = -1
	m.Selected.index = -1
	m.Selected.name = ""
}

// helper resets all selection variables to -1/default
func resetOptions(m *model) {
	resetChannelMode(m)
	resetSelected(m)
}

// helper toggles messages on/off
func showMessages(m *model) {
	m.ShowMessage = !m.ShowMessage
	m.Message = "messages on"
}

// helper clears last message
func clearMessages(m *model) {
	m.Message = fmt.Sprintf("")
}

// move up one device
func cursorUp(m *model, msg tea.Msg) {
	resetChannelMode(m)
	if m.Cursor.pos > 0 {
		m.Cursor.pos--
		clearMessages(m)
	}
	if m.Cursor.pos < (m.Paginator.Page * m.Paginator.PerPage) {
		m.Paginator.PrevPage()
		m.Cursor.pos = (m.Paginator.Page*m.Paginator.PerPage + (m.Paginator.PerPage - 1))
		return
	}
	return
}

// move down one device
func cursorDown(m *model, msg tea.Msg) {
	resetChannelMode(m)
	if m.Cursor.pos < ((m.Count.total - m.Count.cards) - 1) { // exclude cards
		m.Cursor.pos++
		clearMessages(m)
	}
	if m.Cursor.pos >= ((m.Paginator.Page + 1) * m.Paginator.PerPage) {
		m.Paginator.NextPage()
		m.Cursor.pos = (m.Paginator.Page * m.Paginator.PerPage)
		return
	}
	return
}

// go to first device
func cursorFirst(m *model) {
	resetChannelMode(m)
	m.Paginator.Page = 0
	m.Message = fmt.Sprintf("go to first device")
	m.Cursor.pos = 0
	return
}

// go to last device
func cursorLast(m *model) {
	resetChannelMode(m)
	m.Paginator.Page = m.Paginator.TotalPages - 1
	m.Message = fmt.Sprintf("go to last device")
	m.Cursor.pos = (m.Count.total - m.Count.cards) - 1 // exclude cards
	return
}
func cursorPage(m *model, next bool) {
	resetChannelMode(m)
	if next {
		m.Paginator.NextPage()
		m.Message = "next page"
	} else {
		m.Paginator.PrevPage()
		m.Message = "prev page"
	}
	m.Cursor.pos = (m.Paginator.Page * m.Paginator.PerPage)
	return
}

// iterate through channel control
func changeChannel(m *model) {
	if m.ChannelMode < m.Device[m.Cursor.pos].pulsecount-1 {
		m.ChannelMode++
	} else {
		m.ChannelMode = -1
	}
	if m.ChannelMode < 0 { // don't access a negative index
		m.Message = fmt.Sprintf("all channels selected balance: %v", m.Device[m.Cursor.pos].pulsebalance)
		return
	}
	m.Message = fmt.Sprintf("channel: %v balance: %v", m.Device[m.Cursor.pos].pulsechannels[m.ChannelMode], m.Device[m.Cursor.pos].pulsebalance)
}

// toggle fullscreen
func toggleFullscreen(m *model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.Fullscreen {
		cmd = tea.ExitAltScreen
		m.Message = "inline"
	} else {
		cmd = tea.EnterAltScreen
		m.Message = "fullscreen"
	}
	m.Fullscreen = !m.Fullscreen
	return m, cmd
}

// adjust all variables that rely on knowing the size of the terminal
func resizeProgram(m *model, msg tea.WindowSizeMsg) {
	// pass model new terminal size attributes
	m.Width = msg.Width - 4 // set app width a little less than terminal width
	m.Height = msg.Height   // TODO may need in the future (only includes app height)
	// center the application
	if m.Width > setWidth { // fill the margin with empty space if terminal
		m.Width = setWidth // is larger than the configured app width and
		m.Margin = ((msg.Width - m.Width) / 2) - 1
	} else { // make left and right margin equal or....
		m.Margin = 1 // set the margin to a single column otherwise
	}
	// text and bars should be 3/4 the width of the app
	m.StringLen = (m.Width / 4) * 3
	// resize bars based on new width
	for i := 0; i < (m.Count.total - m.Count.cards); i++ {
		for j := 0; j < m.Device[i].pulsecount; j++ {
			m.Device[i].bar[j].Width = m.StringLen
		}
	}
	// resize border based on new width
	m.Border.MarginLeft(m.Margin).Width(m.Width).
		MarginRight(m.Margin)
	// resize help model based on new width
	m.Help.Width = m.Width
}

// handle removal of devices
func refreshPosition(m *model) {
	// m.Paginator.SetTotalPages(len(m.Device))
	m.Paginator.SetTotalPages(m.Count.total - m.Count.cards)
	// handle removal of devices >> reduction in pages
	for m.Paginator.Page > m.Paginator.TotalPages-1 {
		m.Paginator.Page -= 1
		if m.Paginator.Page == 0 {
			break
		}
	}
	// set cursor to last entry if it is set past last actual entry
	if m.Cursor.pos > ((m.Count.total - m.Count.cards) - 1) {
		m.Cursor.pos = ((m.Count.total - m.Count.cards) - 1)
	}
}

// handle minimum terminal size
func validateTerminalSize(m *model, msg tea.WindowSizeMsg) tea.Cmd {
	if msg.Width < minWidth || msg.Height < minHeight {
		cmd := tea.Quit  // next update command exit quit program
		errorHalt = true // prevent the next View() render to stop crashes
		m.Message = fmt.Sprintf("%v exit. Window < %vx%v. \n", programName,
			minWidth, minHeight)
		return cmd
	}
	return nil
}

// show a message on startup if config file did not load properly
func displayInitMessage(m *model) {
	if !isConfigured {
		m.Message = errorLoad
		isConfigured = true
	}
}

func displayProgramMessage(m *model) {
	m.Message = programMsg
}
