// /////////////////////////////////////////////////////////////////////////////
// POPULATE AND CONTROL PULSEDEVICE STRUCTS
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	"fmt"                                       // format and print text
	"github.com/charmbracelet/bubbles/progress" // render progress bars
	tea "github.com/charmbracelet/bubbletea"    // main cli application library
	"github.com/charmbracelet/lipgloss"         // style application
	"os/exec"                                   // run external system commands
	"strconv"                                   // convert types to/from string
	"strings"                                   // manipulate strings
)

// Primary function to set pulsedevice state
func updateDevices(m model) tea.Cmd {
	w := m.StringLen
	d, c := buildDevices(buildPulse())               // devices and device count
	formatProgressBars(d, colorBars(deviceColor), w) // populate/style progress bars of devices
	return func() tea.Msg {                          // new PulseDevice structs
		return RefreshMsg{d, c} // and count on every update
	}
}

// pull progress bar color string from lipgloss adaptive color value
func colorBars(c []lipgloss.AdaptiveColor) []string {
	var colors []string   // progress.WithSolidFill
	for _, v := range c { // only takes the color string
		if lipgloss.HasDarkBackground() { // of a lipgloss style
			colors = append(colors, v.Dark)
			continue
		}
		colors = append(colors, v.Light)
	}
	return colors
}

// populate/style bars for progress model (called by updateDevices())
func formatProgressBars(device []PulseDevice, colors []string, width int) {
	var b Bar                      // create progress bars
	w := progress.WithWidth(width) // assign width and color
	if setNoColor {
		for i := range colors {
			colors[i] = ""
		}
	}
	b.sink = progress.New(progress.WithSolidFill(colors[0]), w)
	b.stream = progress.New(progress.WithSolidFill(colors[1]), w)
	b.source = progress.New(progress.WithSolidFill(colors[2]), w)
	b.output = progress.New(progress.WithSolidFill(colors[3]), w)
	c := lipgloss.NewStyle().Foreground(toggleColor[0])
	b.sink.PercentageStyle = c   // style percentage text
	b.stream.PercentageStyle = c // to inactive color by default
	b.source.PercentageStyle = c
	b.output.PercentageStyle = c

	if hidePercentage == true {
		b.sink.ShowPercentage = false
		b.stream.ShowPercentage = false
		b.source.ShowPercentage = false
		b.output.ShowPercentage = false
	}
	// style bars by device type
	for i := 0; i < len(device); i++ {
		if device[i].pulsetype == 0 {
			for j := 0; j < device[i].pulsecount; j++ {
				device[i].bar = append(device[i].bar, b.sink)
			}
		} else if device[i].pulsetype == 1 {
			for j := 0; j < device[i].pulsecount; j++ {
				device[i].bar = append(device[i].bar, b.stream)
			}
		} else if device[i].pulsetype == 2 {
			for j := 0; j < device[i].pulsecount; j++ {
				device[i].bar = append(device[i].bar, b.source)
			}
		} else if device[i].pulsetype == 3 {
			for j := 0; j < device[i].pulsecount; j++ {
				device[i].bar = append(device[i].bar, b.output)
			}
		}
	}
}

// DEVICE CONTROL
// //////////////////////////////////////////////////////////////////////////////
// mute/unmute a device
func toggleDeviceMute(m *model) {
	var c, r string
	var d int
	switch m.Device[m.Cursor.pos].pulsetype {
	case pulsesink:
		c = sink_mute_cmd
		r = sink_mute_rpl
		d = pulsesink
	case pulsestream:
		c = stream_mute_cmd
		d = pulsestream
	case pulsesource:
		c = source_mute_cmd
		r = source_mute_rpl
		d = pulsesource
	case pulseoutput:
		c = output_mute_cmd
		d = pulseoutput
	}
	num := m.Device[m.Cursor.pos].pulseindex
	index := strconv.Itoa(num)
	cmd := exec.Command(pactl, c, index, toggle)
	err := cmd.Run()
	if err != nil {
		m.Message = fmt.Sprintf("error toggling device mute")
	}
	if d == pulsestream || d == pulseoutput { // no get-mute command for streams/outputs
		m.Message = fmt.Sprintf("mute toggled: %v", m.Device[m.Cursor.pos].pulsedescription)
		return
	}
	out, err := exec.Command(pactl, r, index).Output()
	if err != nil {
		m.Message = fmt.Sprintf("error retrieving device mute")
		fmt.Println("error")
	}
	if strings.Contains(string(out), "yes") {
		m.Message = fmt.Sprintf("muted: %v", m.Device[m.Cursor.pos].pulsedescription)
	} else if strings.Contains(string(out), "no") {
		m.Message = fmt.Sprintf("unmuted: %v", m.Device[m.Cursor.pos].pulsedescription)
	}
}

// terminate parent process of stream, effectiveness depends on target
func killStreamOrOutput(m *model) {
	if m.Device[m.Cursor.pos].pulsetype == pulsesink || m.Device[m.Cursor.pos].pulsetype == pulsesource {
		m.Message = fmt.Sprintf("cannot kill non-stream %v", m.Device[m.Cursor.pos].pulsename)
		return
	}
	if m.Device[m.Cursor.pos].pulsedriver == loopback_c {
		killLoopback(m) // unload if stream is loopback
		return
	}
	proc := m.Device[m.Cursor.pos].pulsepid
	cmd := exec.Command("kill", "-9", proc)
	err := cmd.Run()
	if err != nil {
		m.Message = fmt.Sprintf("error killing stream: %v", m.Device[m.Cursor.pos].pulsename)
	}
	m.Message = fmt.Sprintf("killed stream: %v", m.Device[m.Cursor.pos].pulsename)
}

// make the sink on cursor the default and route all streams to it
func changeDefaultSink(m *model) {
	if m.Selected.devicetype > -1 {
		return
	}
	if m.Device[m.Cursor.pos].pulsetype != pulsesink {
		return
	}
	num := m.Device[m.Cursor.pos].pulseindex
	p := default_sink_cmd
	index := strconv.Itoa(num)
	cmd := exec.Command(pactl, p, index)
	err := cmd.Run()
	if err != nil {
		m.Message = "error changing default sink"
		return
	}
	moveAllStreams(m, getStreamIndexes(m), index) // get streams, move each stream to new default sink
	m.Message = fmt.Sprintf("changed default sink to: %v", m.Device[m.Cursor.pos].pulsename)
	resetSelected(m) // unset Selected after operation
}

// make the source on cursor the default source
func changeDefaultSource(m *model) {
	if m.Selected.devicetype > -1 {
		return
	}
	if m.Device[m.Cursor.pos].pulsetype != pulsesource {
		return
	}
	num := m.Device[m.Cursor.pos].pulseindex
	p := default_source_cmd
	index := strconv.Itoa(num)
	cmd := exec.Command(pactl, p, index)
	err := cmd.Run()
	if err != nil {
		m.Message = "error changing default source"
		return
	}
	m.Message = fmt.Sprintf("changed default source to: %v", m.Device[m.Cursor.pos].pulsename)
	// TODO is migration required here?
	resetSelected(m) // unset Selected after operation
}

// move each stream to a sink (called by changeDefaultSink() )
func moveAllStreams(m *model, stream []int, sink string) {
	for _, v := range stream {
		target := strconv.Itoa(v)
		cmd := exec.Command(pactl, move_stream_cmd, target, sink)
		err := cmd.Run()
		if err != nil {
			m.Message = "error moving streams"
		}
	}
}

// assemble a slice containing each stream (called by moveAllStreams() )
func getStreamIndexes(m *model) []int {
	var index []int
	for _, v := range m.Device {
		if v.pulsetype == pulsestream {
			index = append(index, v.pulseindex)
		}
	}
	return index
}

// set volume on target device, takes array of channel strings
func changeDeviceVolume(m *model, vol []string) {
	var p string
	switch m.Device[m.Cursor.pos].pulsetype {
	case pulsesink:
		p = sink_vol_cmd
	case pulsestream:
		p = stream_vol_cmd
	case pulsesource:
		p = source_vol_cmd
	case pulseoutput:
		p = output_vol_cmd
	}
	num := m.Device[m.Cursor.pos].pulseindex
	target := strconv.Itoa(num)
	var args []string
	args = append(args, p)
	args = append(args, target)
	for _, v := range vol {
		args = append(args, v)
	}
	cmd := exec.Command(pactl, args...)
	err := cmd.Run()
	if err != nil {
		m.Message = "error changing device volume"
	}
}

// prepare volume change strings for single or all channels, skip max volume requests
// takes bool for inc/dec (called by changeDeviceVolume())
func formatDeviceVolume(m *model, inc bool) []string {
	var limit bool
	var prefix string
	if inc {
		limit = true
		prefix = "+"
	} else {
		limit = false
		prefix = "-"
	}
	var entry string
	var vol []string
	for i := 0; i < len(m.Device[m.Cursor.pos].pulsevolume); i++ {
		if m.ChannelMode < 0 { // all channels
			if limit == true && m.Device[m.Cursor.pos].pulsevolume[i] >= m.VolumeLimit {
				entry = fmt.Sprintf("%v0%%", prefix)
				vol = append(vol, entry)
				continue
			}
			entry = fmt.Sprintf("%v%v%%", prefix, setVolume)
		} else if m.ChannelMode == i { // specific channels
			if limit == true && m.Device[m.Cursor.pos].pulsevolume[i] >= m.VolumeLimit {
				entry = fmt.Sprintf("%v0%%", prefix)
				vol = append(vol, entry)
				continue
			}
			entry = fmt.Sprintf("%v%v%%", prefix, setVolume)
		} else { // no changes for excluded channels
			entry = fmt.Sprintf("%v0%%", prefix)
		}
		vol = append(vol, entry)
	}
	return vol
}

// set volume for all channels on target device from 10%-100%
func normalizeDeviceVolume(m *model, v int) {
	var args []string
	var p string
	switch m.Device[m.Cursor.pos].pulsetype {
	case pulsesink:
		p = sink_vol_cmd
	case pulsestream:
		p = stream_vol_cmd
	case pulsesource:
		p = source_vol_cmd
	case pulseoutput:
		p = output_vol_cmd
	}
	args = append(args, p)
	num := m.Device[m.Cursor.pos].pulseindex
	target := strconv.Itoa(num)
	args = append(args, target)
	vol := fmt.Sprintf("%v%%", v)
	args = append(args, vol)
	cmd := exec.Command(pactl, args...)
	err := cmd.Run()
	if err != nil {
		m.Message = "error normalize volume"
	}
	m.ChannelMode = -1 // return to controlling all channels
	m.Message = fmt.Sprintf("volume set to %v%%", v)
}

// move target stream to the selected sink
func moveStreamToSink(m *model) {
	if m.Selected.devicetype != pulsesink {
		m.Message = "enter: move stream to target sink"
		return
	}
	if m.Device[m.Cursor.pos].pulsetype != pulsestream {
		m.Message = "not a valid stream"
		return
	}
	stream := strconv.Itoa(m.Device[m.Cursor.pos].pulseindex)
	sink := strconv.Itoa(m.Selected.index)
	cmd := exec.Command(pactl, move_stream_cmd, stream, sink)
	err := cmd.Run()
	if err != nil {
		m.Message = "error migrating stream"
		return
	}
	m.Message = fmt.Sprintf("stream: #%v sent to sink: #%v", stream, sink)
	resetSelected(m) // unset Selected after operation
}

// move target output to the selected source
func moveOutputToSource(m *model) {
	if m.Selected.devicetype != pulsesource {
		m.Message = "enter: move output to target source"
		return
	}
	if m.Device[m.Cursor.pos].pulsetype != pulseoutput {
		m.Message = "not a valid output stream"
		return
	}
	output := strconv.Itoa(m.Device[m.Cursor.pos].pulseindex)
	source := strconv.Itoa(m.Selected.index)
	cmd := exec.Command(pactl, move_output_cmd, output, source)
	err := cmd.Run()
	if err != nil {
		m.Message = "error migrating output"
		return
	}
	m.Message = fmt.Sprintf("output: #%v sent to source: #%v", output, source)
	resetSelected(m) // unset Selected after operation
}

// terminate target loopback stream (called by killStream())
func killLoopback(m *model) {
	if m.Device[m.Cursor.pos].pulsedriver != loopback_c {
		m.Message = "not a loopback device"
		return
	}
	module := string(m.Device[m.Cursor.pos].pulsemodule)
	name := m.Device[m.Cursor.pos].pulsedescription
	cmd := exec.Command(pactl, unload_module, module)
	err := cmd.Run()
	if err != nil {
		m.Message = fmt.Sprintf("error unloading module: #%v %v", module, name)
	}
	m.Message = fmt.Sprintf("killed %v", name)
}

// unload All loopback streams
func unloadLoopback(m *model) {
	cmd := exec.Command(pactl, unload_module, loopback_module)
	err := cmd.Run()
	if err != nil {
		m.Message = "error unloading loopback module"
	}
	m.Message = "killed all loopback streams"
	resetSelected(m) // unset Selected after operation
}

// loopback a source to a sink
func loopbackSourceToSink(m *model) {
	if m.Selected.devicetype != pulsesink {
		m.Message = "enter: loop source to target sink"
		return
	}
	if m.Device[m.Cursor.pos].pulsetype != pulsesource {
		m.Message = "not a valid source to loopback"
		return
	}
	index := m.Device[m.Cursor.pos].pulseindex
	source := strconv.Itoa(index)
	source = fmt.Sprintf("source=%v", source)
	sink := strconv.Itoa(m.Selected.index)
	sink = fmt.Sprintf("sink=%v", sink)
	latency := strconv.Itoa(varLatency)
	latency = fmt.Sprintf("latency_msec=%v", latency)
	cmd := exec.Command(pactl, load_module, loopback_module, latency, sink, source)
	err := cmd.Run()
	if err != nil {
		m.Message = "error executing source loopback"
		return
	}
	m.Message = fmt.Sprintf("source: #%v sent to sink: #%v", source, sink)
	resetSelected(m) // unset Selected after operation
}

// increment or decrement latency value to be used with pactl commands
func changeLatency(m *model, inc bool) {
	if inc {
		varLatency += incLatency
		if varLatency > maxLatency {
			varLatency = minLatency
		}
	} else {
		varLatency -= incLatency
		if varLatency < minLatency {
			varLatency = maxLatency
		}
	}
	m.Message = fmt.Sprintf("latency set to: %v milliseconds", varLatency)
}

// select the device to subsequently perform an action
func selectDevice(m *model) {
	m.Selected.devicetype = m.Device[m.Cursor.pos].pulsetype
	m.Selected.index = m.Device[m.Cursor.pos].pulseindex
	m.Selected.name = m.Device[m.Cursor.pos].pulsedescription
	m.Message = fmt.Sprintf("device selected")
}

func suspendSinkOrSource(m *model) {
	if m.Selected.devicetype == pulsestream {
		return
	}
	if m.Selected.devicetype == pulseoutput {
		return
	}
	if m.Selected.index != m.Device[m.Cursor.pos].pulseindex {
		return
	}
	d := m.Device[m.Cursor.pos].pulsetype
	var p string
	switch d {
	case 0:
		p = sus_sink_cmd
	case 2:
		p = sus_source_cmd
	}
	target := strconv.Itoa(m.Device[m.Cursor.pos].pulseindex)
	state := "0"
	if m.Device[m.Cursor.pos].pulsestate == running_state {
		state = "1"
	}
	cmd := exec.Command(pactl, p, target, state)
	err := cmd.Run()
	if err != nil {
		m.Message = fmt.Sprintf("error suspending device")
	}
	m.Message = fmt.Sprintf("suspend: %v %v", state, m.Device[m.Cursor.pos].pulsedescription)
	resetSelected(m)
}

// return type of device
func getDeviceType(t int) string {
	var s string
	switch t {
	case 0:
		s = "sink"
	case 1:
		s = "stream"
	case 2:
		s = "source"
	case 3:
		s = "output"
	case 4:
		s = "card"
	}
	return s
}

// TODO fill cases with desired commands
// run pactl command based on the selected device type and target device type
func performAction(m *model) {
	target := m.Device[m.Cursor.pos].pulsetype
	if m.Selected.devicetype < 0 { // ON ENTER KEY PRESS  no device toggled
		switch target {
		case 0: // sink
			changeDefaultSink(m)
		case 1: // stream
			m.Message = "stream operation"
		case 2: // source
			changeDefaultSource(m)
		case 3: // output
			m.Message = "output operation"
		}
	} else if m.Selected.devicetype == 0 { // ON ENTER KEY PRESS with sink toggled
		switch target {
		case 0: // sink
			suspendSinkOrSource(m)
		case 1: // stream
			moveStreamToSink(m)
		case 2: // source
			loopbackSourceToSink(m)
		case 3: // output
			// m.Message = "sink targeting output operation"
		}
	} else if m.Selected.devicetype == 2 { // ON ENTER KEY PRESS with source toggled
		switch target {
		case 0: // sink
			m.Message = "source targeting sink operation"
		case 1: // stream
			m.Message = "source targeting stream operation"
		case 2: // source
			suspendSinkOrSource(m)
		case 3: // output
			moveOutputToSource(m)
			// m.Message = "source targeting output operation"
		}
	}
}
