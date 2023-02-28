// /////////////////////////////////////////////////////////////////////////////
// REQUEST AND PARSE DATA FROM PACTL
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	"bufio"         // read/write input/output
	"encoding/json" // decode json data streams
	"fmt"           // format and print text
	"os/exec"       // run external system commands
	"strconv"       // convert types to/from string
	"strings"       // manipulate strings
)

type Pulse struct {
	Index       int                    `json:"index"`
	Driver      string                 `json:"driver"`
	Module      interface{}            `json:"owner_module"`
	State       string                 `json:"state"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	SinkIndex   int                    `json:"sink"`
	SourceIndex int                    `json:"source"`
	SampleRate  string                 `json:"sample_specification"`
	Channels    string                 `json:"channel_map"`
	Mute        bool                   `json:"mute"`
	Balance     float64                `json:"balance"`
	Volume      map[string]interface{} `json:"volume"`
	Port        string                 `json:"active_port"`
	Latency     float64                `json:"source_latency_usec"`
	Properties  struct {
		Icon         string `json:"application.icon_name"`
		Title        string `json:"media.name"`
		Name         string `json:"application.name"`
		PID          string `json:"application.process.id"`
		Binary       string `json:"application.process.binary"`
		Card         string `json:"alsa.card_name"`
		Bus          string `json:"device.bus"`
		PDescription string `json:"device.description"`
		DevString    string `json:"device.string"`
		Battery      string `json:"bluetooth.battery"`
	} `json:"properties"`
	ChannelList    []string  // individual channel names
	ChannelVolume  []float64 // individual channel values
	FormattedTitle string    // issue #1310 in pactl repo: utf8 characters return "(null)"
}

// return attributes
func (p Pulse) getIndex() int             { return p.Index }
func (p Pulse) getDriver() string         { return p.Driver }
func (p Pulse) getState() string          { return p.State }
func (p Pulse) getName() string           { return p.Name }
func (p Pulse) getDescription() string    { return p.Description }
func (p Pulse) getPDescription() string   { return p.Properties.PDescription }
func (p Pulse) getBus() string            { return p.Properties.Bus }
func (p Pulse) getBattery() string        { return p.Properties.Battery }
func (p Pulse) getDevString() string      { return p.Properties.DevString }
func (p Pulse) getSampleRate() string     { return p.SampleRate }
func (p Pulse) getSinkIndex() int         { return p.SinkIndex }
func (p Pulse) getSourceIndex() int       { return p.SourceIndex }
func (p Pulse) getMute() bool             { return p.Mute }
func (p Pulse) getBalance() float64       { return p.Balance }
func (p Pulse) getPort() string           { return p.Port }
func (p Pulse) getLatency() float64       { return p.Latency }
func (p Pulse) getIconName() string       { return p.Properties.Icon }
func (p Pulse) getAppName() string        { return p.Properties.Name }
func (p Pulse) getBinaryName() string     { return p.Properties.Binary }
func (p Pulse) getPID() string            { return p.Properties.PID }
func (p Pulse) getTitle() string          { return p.Properties.Title }
func (p Pulse) getFormattedTitle() string { return p.FormattedTitle }
func (p Pulse) getChannelCount() int      { return len(p.ChannelList) }
func (p Pulse) getCardName() string       { return p.Properties.Card }
func (p Pulse) getChannelList() []string {
	var channels []string
	sep := ","
	split := strings.SplitAfter(p.Channels, sep)
	for _, v := range split {
		channels = append(channels, strings.Trim(v, sep))
	}
	return channels
}

// type assertion on Volume json string to get percentages
func (p Pulse) getChannelVolume() []float64 {
	var vol []string
	field := "value_percent"
	for _, v := range p.ChannelList {
		entry := p.Volume[v] // Access Each entry under Volume
		entrymap := entry.(map[string]interface{})
		for k, v := range entrymap { // for each key channel entry map
			switch val := v.(type) { // search for value_percent key
			case string: // and get its value
				if k == field {
					vol = append(vol, val)
				}
			}
		}
	}
	var fvolume []float64
	for _, v := range vol {
		entry := strings.Trim(v, "%")
		value, _ := strconv.ParseFloat(entry, 64)
		fvolume = append(fvolume, value)
	}
	return fvolume
}
func (p Pulse) getModule() string {
	s := strings.Trim(fmt.Sprintf("%v", p.Module), "\"")
	return s
}

// check integrity of json data
func validateJson(jsondata []byte) bool {
	check := json.Valid(jsondata)
	return check
}

// get json bytes and number of indexes for a device type
func getPactlBytes(pulsetype int) ([]byte, int) {
	var cmd []byte
	var err error
	switch pulsetype {
	case 0:
		cmd, err = exec.Command("pactl", "-f", "json", "list", "sinks").Output()
	case 1:
		cmd, err = exec.Command("pactl", "-f", "json", "list", "sink-inputs").Output()
	case 2:
		cmd, err = exec.Command("pactl", "-f", "json", "list", "sources").Output()
	case 3:
		cmd, err = exec.Command("pactl", "-f", "json", "list", "source-outputs").Output()
	case 4:
		cmd, err = exec.Command("pactl", "-f", "json", "list", "cards").Output()
	}
	if err != nil {
		return nil, 0
	}
	count := strings.Count(string(cmd), "\"index\":")
	return cmd, count
}
func getStreamText() string {
	cmd, err := exec.Command("pactl", "-f", "text", "list", "sink-inputs").Output()
	if err != nil {
		return ""
	}
	return string(cmd)
}

// return the media titles for each stream (needed for unicode characters)
func formatStreamText(s string) []string {
	if s == "" {
		return nil
	}
	scanner := bufio.NewScanner(strings.NewReader(s))
	field := "media.name = "
	var names []string
	var formattedNames []string
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), field) {
			names = append(names, scanner.Text())
		}
	}
	if len(names) == 0 {
		return nil
	}
	for i := 0; i < len(names); i++ {
		formatted := strings.TrimSpace(names[i])
		formatted = strings.Trim(formatted, field)
		formatted = strings.TrimSpace(formatted)
		formatted = strings.Trim(formatted, "\"")
		formatted = strings.TrimSpace(formatted)
		formattedNames = append(formattedNames, formatted)
	}
	return formattedNames
}

// generate Pulse structs from pactl json and text data
func buildPulse() ([]Pulse, DeviceCount) {
	var devices []Pulse
	var dc DeviceCount
	types := []int{pulsesink, pulsestream, pulsesource, pulseoutput, pulsecard}
	for _, v := range types {
		var pulsearray []Pulse
		var titles []string
		pactljson, count := getPactlBytes(v)
		if count == 0 || !validateJson(pactljson) {
			continue
		}
		json.Unmarshal([]byte(pactljson), &pulsearray)
		if v == pulsestream {
			pactltext := getStreamText()
			titles = formatStreamText(pactltext)
		}
		for i := 0; i < len(pulsearray); i++ {
			switch v {
			case 0:
				dc.sinks++
			case 1:
				dc.streams++
			case 2:
				dc.sources++
			case 3:
				dc.outputs++
			case 4:
				dc.cards++
			}
			if v != pulsecard {
				pulsearray[i].ChannelList = pulsearray[i].getChannelList()
				pulsearray[i].ChannelVolume = pulsearray[i].getChannelVolume()
			}
			if v == pulsestream {
				// TODO ensure vaild range
				pulsearray[i].FormattedTitle = titles[dc.streams-1]
			}
		}
		devices = append(devices, pulsearray...)
	}
	dc.total = dc.sinks + dc.streams + dc.sources + dc.outputs + dc.cards
	return devices, dc
}

// generate complete PulseDevice slice from relevant categories for device types
func buildDevices(p []Pulse, d DeviceCount) ([]PulseDevice, DeviceCount) {
	var devices []PulseDevice
	var index = 0

	for i := 0; i < d.total; i++ {
		devices = append(devices, PulseDevice{})
	}
	for i := index; i < d.sinks; i++ {
		devices[i].pulsetype = pulsesink
		devices[i].pulseindex = p[i].getIndex()
		devices[i].pulsedriver = p[i].getDriver()
		devices[i].pulsemodule = p[i].getModule()
		devices[i].pulsestate = p[i].getState()
		devices[i].pulsename = p[i].getName()
		devices[i].pulsedescription = p[i].getDescription()
		devices[i].pulsesamplerate = p[i].getSampleRate()
		devices[i].pulsecount = p[i].getChannelCount()
		devices[i].pulsechannels = p[i].getChannelList()
		devices[i].pulsevolume = p[i].getChannelVolume()
		devices[i].pulsecard = p[i].getCardName()
		devices[i].pulsemute = p[i].getMute()
		devices[i].pulsebalance = p[i].getBalance()
		devices[i].pulseport = p[i].getPort()
		devices[i].pulsebus = p[i].getBus()
		devices[i].pulsebattery = p[i].getBattery()
		devices[i].pulsedevstring = p[i].getDevString()
		index++
	}
	for i := index; i < d.sinks+d.streams; i++ {
		devices[i].pulsetype = pulsestream
		devices[i].pulseindex = p[i].getIndex()
		devices[i].pulsesinkindex = p[i].getSinkIndex()
		devices[i].pulsedriver = p[i].getDriver()
		devices[i].pulsemodule = p[i].getModule()
		devices[i].pulsename = p[i].getBinaryName()
		devices[i].pulsedescription = p[i].getFormattedTitle()
		devices[i].pulsecount = p[i].getChannelCount()
		devices[i].pulsechannels = p[i].getChannelList()
		devices[i].pulsevolume = p[i].getChannelVolume()
		devices[i].pulsemute = p[i].getMute()
		devices[i].pulsebalance = p[i].getBalance()
		devices[i].pulsepid = p[i].getPID()
		index++
	}
	for i := index; i < d.sinks+d.streams+d.sources; i++ {
		devices[i].pulsetype = pulsesource
		devices[i].pulseindex = p[i].getIndex()
		devices[i].pulsedriver = p[i].getDriver()
		devices[i].pulsemodule = p[i].getModule()
		devices[i].pulsestate = p[i].getState()
		devices[i].pulsename = p[i].getName()
		devices[i].pulsecard = p[i].getCardName()
		devices[i].pulsedescription = p[i].getDescription()
		devices[i].pulsesamplerate = p[i].getSampleRate()
		devices[i].pulsecount = p[i].getChannelCount()
		devices[i].pulsechannels = p[i].getChannelList()
		devices[i].pulsevolume = p[i].getChannelVolume()
		devices[i].pulsemute = p[i].getMute()
		devices[i].pulsebalance = p[i].getBalance()
		devices[i].pulseport = p[i].getPort()
		devices[i].pulsebus = p[i].getBus()
		devices[i].pulsebattery = p[i].getBattery()
		devices[i].pulsedevstring = p[i].getDevString()
		index++
	}
	for i := index; i < d.sinks+d.streams+d.sources+d.outputs; i++ {
		devices[i].pulsetype = pulseoutput
		devices[i].pulseindex = p[i].getIndex()
		devices[i].pulsesourceindex = p[i].getSourceIndex()
		devices[i].pulsedriver = p[i].getDriver()
		devices[i].pulsemodule = p[i].getModule()
		devices[i].pulsesamplerate = p[i].getSampleRate()
		devices[i].pulsename = p[i].getTitle()
		devices[i].pulsedescription = p[i].getIconName()
		devices[i].pulsecount = p[i].getChannelCount()
		devices[i].pulsechannels = p[i].getChannelList()
		devices[i].pulsevolume = p[i].getChannelVolume()
		devices[i].pulsemute = p[i].getMute()
		devices[i].pulsebalance = p[i].getBalance()
		devices[i].pulselatency = p[i].getLatency()
		devices[i].pulsepid = p[i].getPID()
		index++
	}
	for i := index; i < d.sinks+d.streams+d.sources+d.outputs+d.cards; i++ {
		devices[i].pulsetype = pulsecard
		devices[i].pulseindex = p[i].getIndex()
		devices[i].pulsemodule = p[i].getModule()
		devices[i].pulsedescription = p[i].getPDescription()
		devices[i].pulsebattery = p[i].getBattery()
		index++
	}
	return devices, d
}
