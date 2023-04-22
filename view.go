// /////////////////////////////////////////////////////////////////////////////
// BUBBLETEA VIEW
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	"fmt"                                       // format and print text
	"github.com/charmbracelet/bubbles/progress" // render progress bars
	"github.com/charmbracelet/lipgloss"         // style application
	"strconv"                                   // convert types to/from string
	"strings"                                   // manipulate strings
)

// //////////////////////////////////////////////////////////////////////////////
func (m model) View() string {
	if m.Width < minWidth || m.Height < minHeight {
		return "terminal too small"
	}
	if errorHalt { // app in process of exiting, don't try to render crashable code
		return ""
	}
	// help/lipgloss must have latest width from update before rendering view
	// use one lipgloss style to edit text attributes in view using Unset when needed
	// style := lip.Width(m.Width).Align(center).Foreground(toggleColor[1])
	m.Text = lip.Width(m.Width).Align(center).Foreground(toggleColor[1])
	////////////////////////////////////////////////////////////////////////////////
	// Logic Required To Produce View
	// We don't want to show cards in view, so use a slice without them
	// then get slice range for current page and iterate through/display below
	DevicesExcludingCards := m.Device[:m.Count.total-m.Count.cards]
	start, end := m.Paginator.GetSliceBounds(len(DevicesExcludingCards))
	// when multiple elements are rendered on the same line, their widths need to
	// account for one another (in this instance its paginator text and message text)
	// also account for padding that elements use (padding at line start and end)
	paginatorLen := m.Paginator.TotalPages + (len(pad) * 2)
	// paginatorLen := m.Paginator.TotalPages
	pWidth := (m.Width - paginatorLen)
	// update help model text rendering according to the latest width update
	// width set in update function
	helpText := m.Help.View(m.Keys)
	////////////////////////////////////////////////////////////////////////////////
	// exit view if there are no devices to report
	if len(DevicesExcludingCards) < 1 { // change value to debug
		return m.Border.Render(displayEmptyList(&m))
	}
	// Rendered View String
	////////////////////////////////////////////////////////////////////////////////
	s := ""
	if !setNoTitle { // program header can be toggled off in config
		s += m.Text.Render(fmt.Sprintf("%v", programName)) // one line
		s += "\n"                                          // one line
	}
	// TODO use/remove debug strings
	// s += pad + m.Text.UnsetAlign().Render(fmt.Sprintf("m.Display.level: %v", m.Display.level))

	// show current selected device if any // one line
	s += pad + m.Text.UnsetAlign().Render(cutText(displayToggledDevice(&m), m.StringLen / 2))
	s += "\n\n" // two lines
	// loop through each device and add its channel info to the view string
	for index, pulsedevice := range DevicesExcludingCards[start:end] {
		s += displayEntry(&m, pulsedevice, index)
	}
	s += pad + fmt.Sprintf("%v", m.Paginator.View()) // paginator counter +
	if m.ShowMessage {                               // messages rendered on the same line
		m.Text.UnsetForeground()
		m.Text.Foreground(toggleColor[1])
		// s += m.Text.Width(pWidth).Align(center).Render(cutText(fmt.Sprintf("%v", m.Message), m.StringLen))
		s += m.Text.Width(pWidth).Align(right).Render(cutText(fmt.Sprintf("%v", m.Message), m.StringLen))
	}
	s += "\n\n"     // two lines
	if !setNoHelp { // can be toggled off in config or with flag
		// help module (include a newline while rending please)
		s += m.Text.Align(center).Width(m.Width).Render(helpText + "\n") // 1 or 3 lines
	}
	return m.Border.Render(s) // return view inside border
}

// functions used by view()
// //////////////////////////////////////////////////////////////////////////////
// helper function to format the appearance of devices in View(), takes model, device, and position on page
func displayEntry(m *model, d PulseDevice, index int) string {
	var s string
	m.Text = lip.Width(m.Width).Align(center).Foreground(toggleColor[0])
	setChosen(m, d, index) // cursor indicator and variables
	if m.Cursor.chosen {
		m.Text.Foreground(toggleColor[1])
	}
	switch d.pulsetype {
	case 0:
		s += displaySink(m, d)
	case 1:
		s += displayStream(m, d)
	case 2:
		s += displaySource(m, d)
	case 3:
		s += displayOutput(m, d)
	}
	s += "\n"
	s += displayProgressBars(m, d, index)
	return s
}

// helper functions to format the titles for each type of device in displayEntry()
func displaySink(m *model, d PulseDevice) string {
	var s, t1, t2, mute string
	if d.pulsemute {
		mute = muted_state // only used for its string length
	}
	switch m.Display.level {
	case 1:
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + fmt.Sprintf("%v", m.Cursor.suff))
	case 2:
		t1 += cutText(fmt.Sprintf("%v%v%v", displayState(d), displayBattery(getBattery(d)), d.pulsedescription), m.StringLen-(len(m.Cursor.pref)+len(m.Cursor.suff)+len(mute)+len(displayState(d))))
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + t1 + fmt.Sprintf("%v", m.Cursor.suff))
	case 3:
		t1 += cutText(fmt.Sprintf("%v%v", displayBattery(getBattery(d)), d.pulsedescription), m.StringLen)
		t2 += cutText(fmt.Sprintf("%vsink #%v %v %v", displayState(d), d.pulseindex, d.pulsesamplerate, d.pulseport), m.StringLen)
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref)+t1+fmt.Sprintf("%v", m.Cursor.suff)) + "\n"
		s += m.Text.Render(t2)
	}
	return s
}
func displayStream(m *model, d PulseDevice) string {
	var s, t1, t2, mute string
	if d.pulsemute {
		mute = muted_state // only used for its string length
	}
	switch m.Display.level {
	case 1:
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + fmt.Sprintf("%v", m.Cursor.suff))
	case 2:
		t1 += cutText(fmt.Sprintf("%v%v", displayStreamMute(d), d.pulsedescription), m.StringLen-(len(m.Cursor.pref)+len(m.Cursor.suff)+len(mute)+len(displayStreamMute(d))))
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + t1 + fmt.Sprintf("%v", m.Cursor.suff))
	case 3:
		t1 += cutText(fmt.Sprintf("%v", d.pulsedescription), m.StringLen-(len(m.Cursor.pref)+len(m.Cursor.suff)+len(mute)))
		t2 += cutText(fmt.Sprintf("%v%v #%v %v", displayStreamMute(d), d.pulsename, d.pulseindex, displaySinkPort(m, d.pulsesinkindex)), m.StringLen)
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref)+t1+fmt.Sprintf("%v", m.Cursor.suff)) + "\n"
		s += m.Text.Render(t2)
	}
	return s
}
func displaySource(m *model, d PulseDevice) string {
	var s, t1, t2, mute string
	if d.pulsemute {
		mute = muted_state // only used for its string length
	}
	switch m.Display.level {
	case 1:
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + fmt.Sprintf("%v", m.Cursor.suff))
	case 2:
		t1 += cutText(fmt.Sprintf("%v%v%v", displayState(d), displayBattery(getBattery(d)), d.pulsedescription), m.StringLen-(len(m.Cursor.pref)+len(m.Cursor.suff)+len(mute)+len(displayState(d))))
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + t1 + fmt.Sprintf("%v", m.Cursor.suff))
	case 3:
		t1 += cutText(fmt.Sprintf("%v%v", displayBattery(getBattery(d)), d.pulsedescription), m.StringLen-(len(m.Cursor.pref)+len(m.Cursor.suff)+len(mute)))
		t2 += cutText(fmt.Sprintf("%vsource #%v %v %v", displayState(d), d.pulseindex, d.pulsesamplerate, d.pulseport), m.StringLen)
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref)+t1+fmt.Sprintf("%v", m.Cursor.suff)) + "\n"
		s += m.Text.Render(t2)
	}
	return s
}
func displayOutput(m *model, d PulseDevice) string {
	var s, t1, t2, mute string
	if d.pulsemute {
		mute = muted_state // only used for its string length
	}
	switch m.Display.level {
	case 1:
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + fmt.Sprintf("%v", m.Cursor.suff))
	case 2:
		t1 += cutText(fmt.Sprintf("%v%v", displayOutputMute(d), d.pulsename), m.StringLen-(len(m.Cursor.pref)+len(m.Cursor.suff)+len(mute)+len(displayOutputMute(d))))
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref) + t1 + fmt.Sprintf("%v", m.Cursor.suff))
	case 3:
		t1 += cutText(fmt.Sprintf("%v", d.pulsename), m.StringLen-(len(m.Cursor.pref)+len(m.Cursor.suff)+len(mute)))
		t2 += cutText(fmt.Sprintf("%v#%v %v %v %vµs", displayOutputMute(d), d.pulseindex, displaySourceName(m, d.pulsesourceindex), d.pulsesamplerate, d.pulselatency), m.StringLen)
		s += m.Text.Render(fmt.Sprintf("%v", m.Cursor.pref)+t1+fmt.Sprintf("%v", m.Cursor.suff)) + "\n"
		s += m.Text.Render(t2)
	}
	return s
}

// helper function to render progress bars in View()
func displayProgressBars(m *model, d PulseDevice, index int) string {
	var s string
	var chosen bool
	page := m.Paginator.Page * m.Paginator.PerPage
	if index+page == m.Cursor.pos { // determine color
		chosen = true
	}
	for i := 0; i < d.pulsecount; i++ {
		if m.ChannelMode == i && chosen {
			m.Text.Width(m.Width).Align(center).Foreground(toggleColor[1])
			var selected string // get adaptive color's string for progress
			selected = toggleColor[1].Dark
			if setNoColor {
				toggleColor[1].Dark = ""
			}
			if !lipgloss.HasDarkBackground() {
				selected = toggleColor[1].Light
				if setNoColor {
					toggleColor[1].Light = ""
				}
			}
			d.bar[i] = progress.New(progress.WithSolidFill(selected), progress.WithWidth(m.StringLen))
			d.bar[i].PercentageStyle = lipgloss.NewStyle().Foreground(toggleColor[1])
			if hidePercentage == true {
				d.bar[i].ShowPercentage = false
			}
			s += m.Text.Render(displayChannel(m, d, index, i)+d.bar[i].ViewAs(float64(d.pulsevolume[i])/float64(100))) + "\n\n"
		} else {
			m.Text.Width(m.Width).Align(center).Foreground(toggleColor[0])
			s += m.Text.Render(displayChannel(m, d, index, i)+d.bar[i].ViewAs(float64(d.pulsevolume[i])/float64(100))) + "\n\n"
		}
	}
	return s
}

// helper function to abbreviate/display channel label based on channel_map string
func displayChannel(m *model, d PulseDevice, indexOnPage int, channel int) string {
	if m.Display.level == 0 {
		return "" // don't show any channel labels
	}
	i := indexOnPage
	c := channel
	switch d.pulsechannels[c] {
	case "mono":
		return displayCursor(m, d, i, c) + "M  "
	case "front-left":
		return displayCursor(m, d, i, c) + "L  "
	case "front-right":
		return displayCursor(m, d, i, c) + "R  "
	case "front-center":
		return displayCursor(m, d, i, c) + "C  "
	case "rear-left":
		return displayCursor(m, d, i, c) + "l  "
	case "rear-right":
		return displayCursor(m, d, i, c) + "r  "
	case "rear-center":
		return displayCursor(m, d, i, c) + "c  "
	default: // auxiliary channels if needed
		return displayCursor(m, d, i, c) + "A  "
		// TODO add cases for top-left/right/center and lfe
	}
}

// helper function to display which channel is selected using changeChannel()
func displayCursor(m *model, d PulseDevice, indexOnPage int, channel int) string {
	cursor := "  "                                 // render nothing but allocate space
	page := m.Paginator.Page * m.Paginator.PerPage // page in current model
	if m.Cursor.pos == indexOnPage+page {          // cursor matches current device
		if m.ChannelMode == channel { // mode matches channel number
			cursor = "> " // render arrow in allocated space
		}
	}
	return cursor // show marker in view
}

// helper function to display the currently selected device if any
func displayToggledDevice(m *model) string {
	var s string
	if m.Selected.devicetype < 0 {
		return s
	}
	s = fmt.Sprintf("%v #%v %v",
		getDeviceType(m.Selected.devicetype), m.Selected.index, m.Selected.name)
	return s
}

// helper function to prevent text from wrapping into newline
func cutText(s string, max int) string {
	trunc := " .."
	if len(s) < max+len(trunc) {
		return s
	}
	return s[:max] + " .."
}

// helper function to get source name from index
func displaySourceName(m *model, index int) string {
	for _, v := range m.Device {
		if v.pulseindex == index {
			return v.pulsecard
		}
	}
	return ""
}

// helper function to get sink port from index
func displaySinkPort(m *model, index int) string {
	for _, v := range m.Device {
		if v.pulseindex == index {
			return fmt.Sprintf("sink #%v %v", v.pulseindex, v.pulseport)
		}
	}
	return ""
}

// helper function to display stream mute state
func displayStreamMute(d PulseDevice) string {
	if d.pulsemute {
		return muted_icon
	}
	return unmuted_icon
}

// helper function to display output mute state
func displayOutputMute(d PulseDevice) string {
	if d.pulsemute {
		return muted_icon
	}
	return mic_icon
}

// helper function to display running state
func displayState(d PulseDevice) string {
	if !istty && !setNoSymbol {
		if d.pulsemute {
			return muted_icon
		}
		if !d.pulsemute && d.pulsestate == running_state {
			return unmuted_icon
		}
		if !d.pulsemute && d.pulsestate == idle_state {
			return idle_icon
		}
		if !d.pulsemute && d.pulsestate == suspended_state {
			return sus_icon
		}
	}
	if d.pulsemute {
		return "(Muted)"
	}
	if !d.pulsemute && d.pulsestate == running_state {
		return running_state + " "
	}
	if !d.pulsemute && d.pulsestate == idle_state {
		return idle_state + " "
	}
	if !d.pulsemute && d.pulsestate == suspended_state {
		return suspended_state + " "
	}
	return ""
}

// helper function returns battery property for pulsedevice
func getBattery(d PulseDevice) string {
	return d.pulsebattery
}

// helper function returns battery percentage icon or percentage string
func displayBattery(status string) string {
	if status == "" {
		return status // no battery property, return empty
	}
	if istty || setNoSymbol {
		return status + " " // return the raw percentage string plus space
	}
	var s string
	b, _ := strconv.Atoi(strings.Trim(status, "%"))
	if b > 89 {
		s = bluetooth_battery_icon[90]
	} else if b > 79 {
		s = bluetooth_battery_icon[80]
	} else if b > 69 {
		s = bluetooth_battery_icon[70]
	} else if b > 59 {
		s = bluetooth_battery_icon[60]
	} else if b > 49 {
		s = bluetooth_battery_icon[50]
	} else if b > 39 {
		s = bluetooth_battery_icon[40]
	} else if b > 29 {
		s = bluetooth_battery_icon[30]
	} else if b > 19 {
		s = bluetooth_battery_icon[20]
	} else if b > 9 {
		s = bluetooth_battery_icon[10]
	} else if b > 0 {
		s = bluetooth_battery_icon[0]
	}
	return s
}

// helper function to bypass rendering logic in view if no devices can be found
func displayEmptyList(m *model) string {
	style := lip.Width(m.Width).Align(center).Foreground(toggleColor[1])
	s := style.Render(fmt.Sprintf("%v", programName))
	s += "\n\n"
	s += style.Render(fmt.Sprintf("No Devices To Report"))
	s += "\n"
	return s
}

// helper function called in displayEntry() to establish the current selected device for highlighting
func setChosen(m *model, d PulseDevice, index int) {
	page := m.Paginator.Page * m.Paginator.PerPage
	if index+page == m.Cursor.pos {
		m.Cursor.chosen = true
		m.Cursor.pref = pref_icon
		m.Cursor.suff = suff_icon
	} else {
		m.Cursor.chosen = false
		m.Cursor.pref = ""
		m.Cursor.suff = ""
	}
}

// helper function to iterate through text verbosity for devices
func changeDisplayLevel(m *model) {
	m.Display.level++
	hidePercentage = false
	if m.Display.level > m.Display.max {
		m.Display.level = 1
		hidePercentage = true
	}
}
