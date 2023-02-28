// /////////////////////////////////////////////////////////////////////////////
// APPLICATION SETTINGS AND STRUCTS
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	"github.com/charmbracelet/bubbles/help"      // manage help messages
	"github.com/charmbracelet/bubbles/key"       // define application key map
	"github.com/charmbracelet/bubbles/paginator" // split application results
	"github.com/charmbracelet/bubbles/progress"  // render progress bars
	tea "github.com/charmbracelet/bubbletea"     // main cli application library
	"github.com/charmbracelet/lipgloss"          // style application
	"strings"                                    // manipulate strings
	"time"                                       // update at interval
)

// Error Variables
var (
	errorHalt = false // stop View() from rendering crashable code
	errorMsg1 = "pulseaudio: \"pactl\" not found in path\nplease install before running program"
	errorMsg2 = "pulseaudio program not found in path \nplease install before running program"
	errorMsg3 = "pulseaudio server process not found\nstart pulseaudio before running program"
	errorLoad string
	errorFmt  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Red)).Padding(0, 1, 0, 1)
	errorOut  = errorFmt.Render
	errorConf = false
)

// Define Colors (it is easier to reference words for the base 16 colors)
const (
	black   = "0"
	red     = "1" // regular colors
	green   = "2"
	yellow  = "3"
	blue    = "4"
	magenta = "5"
	cyan    = "6"
	white   = "7"
	Black   = "8"
	Red     = "9" // bold colors
	Green   = "10"
	Yellow  = "11"
	Blue    = "12"
	Magenta = "13"
	Cyan    = "14"
	White   = "15"
)

// default colors if not defined in config file; light/dark = 0/1 index
var (
	inactiveColor = []string{black, Red}
	activeColor   = []string{red, Magenta}
	sinkColor     = []string{black, White}
	streamColor   = []string{green, blue}
	sourceColor   = []string{black, White}
	outputColor   = []string{green, blue}
)

// apply defined colors to lipgloss adaptive color settting
var (
	lip         = lipgloss.NewStyle()
	toggleColor []lipgloss.AdaptiveColor
	deviceColor []lipgloss.AdaptiveColor
)

// text constants
const (
	programName = "Pulse Manager" // header for application
	programMsg  = "control your pulseaudio devices from the terminal"
	textPad     = 2               // text spacing value
	minWidth    = 45              // minimum width will trigger exit
	minHeight   = 8               // minimum height will trigger exit

	center = lipgloss.Center // alignment for rendering lipgloss text
	left   = lipgloss.Left
	right  = lipgloss.Right
)

// string formatting
var (
	istty          = false                        // is terminal a console? no icons
	isConfigured   = false                        // has model been setup
	hidePercentage = false                        // pass hide percentage to formatProgressBars()
	pad            = strings.Repeat(" ", textPad) // spaces using textPadding
)

// configuration limits (can not be altered)
const (
	minConfigWidth     = 45  // program will exit if less than this value
	maxConfigWidth     = 300 // maximum initial size (can be larger with winSizeMsg)
	minConfigItems     = 1   // set to show at least one item (0 is crashable)
	maxConfigItems     = 12  // 12 is reasonable even at high resolution
	minConfigVolume    = 50  // allows restriction on how high l will set volume
	maxConfigVolume    = 180 // allows for boosting some streams if needed (bad for sinks)
	minConfigIncrement = 1   // smallest percentage increase
	maxConfigIncrement = 30  // allows for large jumps using h/l
)

// pulseaudio device enumeration
const (
	pulsesink   = iota // sink   = 0
	pulsestream        // stream = 1
	pulsesource        // source = 2
	pulseoutput        // output = 3
	pulsecard          // card   = 4
)

// pactl commands
const (
	pactl              = "pactl"
	toggle             = "toggle"
	default_sink_cmd   = "set-default-sink"
	default_source_cmd = "set-default-source"
	sus_sink_cmd       = "suspend-sink"
	sus_source_cmd     = "suspend-source"
	move_stream_cmd    = "move-sink-input"
	move_output_cmd    = "move-source-output"
	load_module        = "load-module"
	unload_module      = "unload-module"
	loopback_module    = "module-loopback"
	sink_vol_cmd       = "set-sink-volume"
	stream_vol_cmd     = "set-sink-input-volume"
	source_vol_cmd     = "set-source-volume"
	output_vol_cmd     = "set-source-output-volume"
	sink_mute_cmd      = "set-sink-mute"
	stream_mute_cmd    = "set-sink-input-mute"
	source_mute_cmd    = "set-source-mute"
	output_mute_cmd    = "set-source-output-mute"
	sink_mute_rpl      = "get-sink-mute"
	source_mute_rpl    = "get-source-mute"
	running_state      = "RUNNING"
	idle_state         = "IDLE"
	suspended_state    = "SUSPENDED"
	muted_state        = "(Muted) "
	loopback_c         = "module-loopback.c"
	bluez5_c           = "module-bluez5-device.c"
	bluetooth          = "bluetooth"
	minLatency         = 10  // minimum latency setting
	maxLatency         = 500 // maximum latency setting
	incLatency         = 10  // latency adjustment amount
)

var varLatency = 10 // pass variable latency value to loopback

// tui icons set by isConsole()
var (
	muted_icon   = "🔇 "
	unmuted_icon = "🔊 "
	idle_icon    = "🔉 "
	sus_icon     = "🔈 "
	volume_icon  = "🎚 "
	mic_icon     = "🎙 "
	pref_icon    = "⮞  "
	suff_icon    = "  ⮜"
	battery_icon = map[int]string{90: " ", 80: " ", 70: " ", 60: " ",
		50: " ", 40: " ", 30: " ", 20: " ", 10: " ", 0: ""}
	//
)

// holds pulseaudio data for model (not all attributes used for each type)
type PulseDevice struct {
	pulsetype        int              // sink, stream, source, output
	pulseindex       int              // index number for sink, stream, source
	pulsesinkindex   int              // index number of destination device for stream
	pulsesourceindex int              // index number of origin device for output
	pulsedriver      string           // driver file used (identify modules with this)
	pulsemodule      string           // module number (target loopback streams)
	pulsestate       string           // running, idle, suspended
	pulsename        string           // sound card name, or stream name
	pulsedescription string           // verbose card name, or stream title
	pulsesamplerate  string           // bits and frequency for sound card
	pulsecount       int              // number of channels (typically 2 for L/R)
	pulsechannels    []string         // channel map strings (front-left, front-right)
	pulsevolume      []float64        // value_percent for each channel
	pulsebalance     float64          // -1.0 to 1.0, 0.0 is balanced
	pulselatency     float64          // source output delay in microseconds
	pulsecard        string           // name of the physical sound card
	pulsemute        bool             // device mute state
	bar              []progress.Model // model is just a struct of data for rendering bar
	pulsepid         string           // process id number to kill a stream
	pulseport        string           // physical port used for sink/source (e.g. mic, line-in)
	pulsebus         string           // physical port used for sink/source (e.g. mic, line-in)
	pulsebattery     string           // bluetooth battery level
	pulsedevstring   string           // find bluetooth devices
}

// track quantities of types in model
type DeviceCount struct {
	sinks   int // number of sinks
	streams int // number of streams
	sources int // number of sources
	outputs int // number of outputs
	cards   int // number of cards
	total   int // total pactl device entries in model
}

const interval = 1000  // program update interval in milliseconds
type TickMsg time.Time // used by bubbletea tea.Tick function
func tickCmd() tea.Cmd { // update program at set interval
	return tea.Tick(interval*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// data that must get sent at interval
type RefreshMsg struct {
	device []PulseDevice // pulseaudio devices
	count  DeviceCount   // the number of each device type
}

// keep track of toggled device type and attributes
type SelectedDevice struct {
	devicetype int
	index      int
	name       string
}

// send device selection settings to bubbletea model
func initSelection() SelectedDevice {
	var s SelectedDevice
	s.devicetype = -1
	s.index = -1
	s.name = ""
	return s
}

// track cursor position and attributes
type Cursor struct {
	pos    int    // position in PulseDevice slice
	chosen bool   // mark the PulseDevice at position as chosen
	pref   string // prefix for item on cursor
	suff   string // suffix for item on cursor
}

// send initial cursor settings to bubbletea model
func initCursor() Cursor {
	var cursor Cursor
	cursor.pref = pref_icon
	cursor.suff = suff_icon
	return cursor
}

// control verbosity of device descriptions
type Display struct {
	level int
	max   int
}

// send initial display settings to bubbletea model
func initDisplay() Display {
	var display Display
	display.level = 1
	display.max = 2
	return display
}

// main bubbletea model
type model struct { // main bubbletea model
	Device      []PulseDevice   // contains PulseDevice structs
	Count       DeviceCount     // number of each type of device
	Keys        programKeymap   // keymaps for program
	Paginator   paginator.Model // manages pagination
	Cursor      Cursor          // displayed position attributes
	ChannelMode int             // control a specific channel
	Width       int             // terminal width
	Margin      int             // margin calculated using terminal width
	Height      int             // terminal height
	Help        help.Model      // manage help messages
	Message     string          // show a helpful message on keypress
	ShowMessage bool            // toggle messages on/off in view
	Fullscreen  bool            // display program fullscreen
	VolumeLimit float64         // limit volume increases
	BarStyle    Bar             // how bar is styled with lipgloss
	StringLen   int             // truncate strings outside of app width
	Border      lipgloss.Style  // how application border is styled
	Text        lipgloss.Style  // how application text is styled
	Selected    SelectedDevice  // which device and what type is selected
	Display     Display         // how much information to show for device
}

// format progress bar by type, copy to pulsedevice
type Bar struct {
	sink   progress.Model // sinks
	stream progress.Model // streams
	source progress.Model // sources
	output progress.Model // outputs
}

// key bindings that will be used in bubbletea model
type programKeymap struct {
	Quit           key.Binding
	Up             key.Binding
	Down           key.Binding
	VolumeUp       key.Binding
	VolumeDown     key.Binding
	VolumeMute     key.Binding
	GoToStart      key.Binding
	GoToEnd        key.Binding
	Mute           key.Binding
	NextPage       key.Binding
	PrevPage       key.Binding
	Escape         key.Binding
	Refresh        key.Binding
	ChangeChannel  key.Binding
	ChangeDisplay  key.Binding
	ShowFullHelp   key.Binding
	KillStream     key.Binding
	UnloadLoopback key.Binding
	Fullscreen     key.Binding
	PerformAction  key.Binding
	SelectDevice   key.Binding
	ShowMessage    key.Binding
	Select         key.Binding
	Volume10       key.Binding
	Volume20       key.Binding
	Volume30       key.Binding
	Volume40       key.Binding
	Volume50       key.Binding
	Volume60       key.Binding
	Volume70       key.Binding
	Volume80       key.Binding
	Volume90       key.Binding
	Volume100      key.Binding
	LatencyUp      key.Binding
	LatencyDown    key.Binding
	Demo           key.Binding
}

// return a pointer to struct of key bindings to be used ↑↓
func initKeymapping() *programKeymap {
	return &programKeymap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k", "cursor up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j", "cursor down"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g", "first entry"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G", "last entry"),
		),
		Mute: key.NewBinding(
			key.WithKeys("m", " "),
			key.WithHelp("m", "mute"),
		),
		VolumeUp: key.NewBinding(
			key.WithKeys("l", "K", "right"),
			key.WithHelp("l", "volume up"),
		),
		VolumeDown: key.NewBinding(
			key.WithKeys("h", "J", "left"),
			key.WithHelp("h", "volume down"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("n", "pgdown"),
			key.WithHelp("n", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("p", "N", "pgup"),
			key.WithHelp("p", "prev page"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		PerformAction: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎", "perform action"),
		),
		ChangeChannel: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "channel"),
		),
		ChangeDisplay: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "change display"),
		),
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		ShowMessage: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "show messages"),
		),
		KillStream: key.NewBinding(
			key.WithKeys("x", "delete"),
			key.WithHelp("x", "kill stream"),
		),
		UnloadLoopback: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("X", "unload loopback"),
		),
		Fullscreen: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fullscreen"),
		),
		SelectDevice: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "select device"),
		),
		Volume10: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "volume 10%"),
		),
		Volume20: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "volume 20%"),
		),
		Volume30: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "volume 30%"),
		),
		Volume40: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "volume 40%"),
		),
		Volume50: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "volume 50%"),
		),
		Volume60: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "volume 60%"),
		),
		Volume70: key.NewBinding(
			key.WithKeys("7"),
			key.WithHelp("7", "volume 70%"),
		),
		Volume80: key.NewBinding(
			key.WithKeys("8"),
			key.WithHelp("8", "volume 80%"),
		),
		Volume90: key.NewBinding(
			key.WithKeys("9"),
			key.WithHelp("9", "volume 90%"),
		),
		Volume100: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "volume 100%"),
		),
		LatencyUp: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "inc latency"),
		),
		LatencyDown: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "dec latency"),
		),
		Demo: key.NewBinding(
			key.WithKeys("d"),
		),
	}
}

// ShortHelp returns keybindings for mini help view in bubbles library
func (k programKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.ShowFullHelp, k.Quit}
}

// FullHelp returns keybindings expanded help view in bubbles library
func (k programKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ChangeChannel},                 // first column
		{k.NextPage, k.PrevPage, k.Escape},              // second column
		{k.VolumeUp, k.VolumeDown, k.Mute},              // third column
		{k.SelectDevice, k.PerformAction, k.KillStream}, // fourth column
		{k.ShowMessage, k.Fullscreen, k.ChangeDisplay},  // fifth column
	}
}

// set paginator values and styles for model, takes item limit and total items
func initPager(limit, total int) paginator.Model {
	p := paginator.New()                                       // initialize new paginator
	p.Type = paginator.Dots                                    // use dots or arabic numbers for page label
	p.PerPage = limit                                          // maximum number of items on each page (flag)
	p.SetTotalPages(total)                                     // calculate max pages based total/PerPage (len)
	p.InactiveDot = lip.Foreground(toggleColor[0]).Render("•") // inactive dot view
	p.ActiveDot = lip.Foreground(toggleColor[1]).Render("•")   // active dot view
	return p
}

// set help model values and styles and send to bubbletea model
func initHelp() help.Model {
	style0 := lipgloss.NewStyle().Foreground(toggleColor[0])
	style1 := lipgloss.NewStyle().Foreground(toggleColor[1])
	h := help.New()
	h.Styles.ShortKey = style0
	h.Styles.FullKey = style0
	h.Styles.ShortDesc = style1
	h.Styles.ShortSeparator = style1
	h.Styles.FullDesc = style1
	h.Styles.FullSeparator = style1
	h.Styles.Ellipsis = style1
	h.FullSeparator = " "
	return h
}

// set lipgloss border style for view in bubbletea model
func initBorder(border lipgloss.Border) lipgloss.Style {
	s := lip.MarginTop(1).MarginBottom(1).
		MarginLeft(0).MarginRight(0).
		Padding(1, 0, 0, 0).Border(border).
		BorderForeground(toggleColor[1])
	return s
}

// set lipgloss text style for view in bubbletea model
func initText() lipgloss.Style {
	s := lipgloss.NewStyle()
	return s
}
