// /////////////////////////////////////////////////////////////////////////////
// CONFIGURATION FILE AND FLAGS
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	// "fmt"                                // format and print text
	"github.com/charmbracelet/lipgloss"  // style application
	flag "github.com/cornfeedhobo/pflag" // command line flag parsing
	"github.com/spf13/viper"             // manage configuration of app
)

// final values passed to tea.Model on Initialization
var (
	setMaxVolume  float64 // maximum volume
	setVolume     float64 // volume adjustment amount
	setNoMessages bool    // program displays messages
	setAltscreen  bool    // program starts fullscreen
	setItems      int     // number of items per page
	setWidth      int     // width of the application - 2 for safety
	setNoColor    bool    // is NO_COLOR set on system?
	setNoHelp     bool    // hide help model
	setNoTitle    bool    // hide title
	setBorder     = lipgloss.NormalBorder()
	setNoSymbol   bool // do not use unicode symbols
	setDisplay    int  // device display level
)

// flag variables used for command line parsing and validation
var (
	configFlag     bool
	fullscreenFlag bool
	messagesFlag   bool
	titleFlag      bool
	helpFlag       bool
	setVolumeFlag  int
	maxVolumeFlag  int
	setItemsFlag   int
	setWidthFlag   int
	symbolsFlag    bool
	displayFlag    int
)

// define the default settings for both flags and config file
func initDefaults() Config {
	var c Config
	c.Settings.Fullscreen = true
	c.Settings.NoMessage = false
	c.Settings.NoHelp = false
	c.Settings.Width = 100
	c.Settings.Items = 4
	c.Settings.VolumeLimit = 110
	c.Settings.VolumeSteps = 5
	c.Settings.NoSymbols = false
	c.Settings.DeviceDisplay = 2
	return c
}

// set the default settings values in viper
func setDefaults(d Config) {
	viper.SetDefault("fullscreen", d.Settings.Fullscreen)
	viper.SetDefault("no-message", d.Settings.NoMessage)
	viper.SetDefault("no-help", d.Settings.NoHelp)
	viper.SetDefault("no-title", d.Settings.NoTitle)
	viper.SetDefault("width", d.Settings.Width)
	viper.SetDefault("items", d.Settings.Items)
	viper.SetDefault("volume-limit", d.Settings.VolumeLimit)
	viper.SetDefault("volume-steps", d.Settings.VolumeSteps)
	viper.SetDefault("no-symbols", d.Settings.NoSymbols)
	viper.SetDefault("device-display", d.Settings.DeviceDisplay)
}

// get color values from configuration file
func readConfig(c *Config) {
	viper.SetConfigName("config.yaml") // name of config file
	viper.SetConfigType("yaml")        // extension
	// viper.AddConfigPath(".")                       // working directory
	viper.AddConfigPath("$HOME/.config/pulsemanager/") // config directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			errorConf = true
		} else {
			// Config file was found but there was another error
			errorConf = true
			errorLoad = "error reading config file"
		}
	}
	if err := viper.Unmarshal(&c); err != nil {
		errorConf = true
		errorLoad = "error reading config file"
		return
	}
}
func validateConfig(c *Config) {
	// use config variable default values if config file settings are not sane
	if c.Settings.Width < minConfigWidth {
		c.Settings.Width = viper.GetInt("width")
	}
	if c.Settings.Width > maxConfigWidth {
		c.Settings.Width = viper.GetInt("width")
	}
	if c.Settings.Items < minConfigItems {
		c.Settings.Items = viper.GetInt("items")
	}
	if c.Settings.Items > maxConfigItems {
		c.Settings.Items = viper.GetInt("items")
	}
	if c.Settings.VolumeLimit < minConfigVolume {
		c.Settings.VolumeLimit = viper.GetInt("volume-limit")
	}
	if c.Settings.VolumeLimit > maxConfigVolume {
		c.Settings.VolumeLimit = viper.GetInt("volume-limit")
	}
	if c.Settings.VolumeSteps < minConfigIncrement {
		c.Settings.VolumeSteps = viper.GetInt("volume-steps")
	}
	if c.Settings.VolumeSteps > maxConfigIncrement {
		c.Settings.VolumeSteps = viper.GetInt("volume-steps")
	}
	if c.Settings.DeviceDisplay < minConfigDisplay {
		c.Settings.DeviceDisplay = viper.GetInt("device-display")
	}
	if c.Settings.DeviceDisplay > maxConfigDisplay {
		c.Settings.DeviceDisplay = viper.GetInt("device-display")
	}
	viper.Set("fullscreen", c.Settings.Fullscreen)
	viper.Set("no-message", c.Settings.NoMessage)
	viper.Set("no-help", c.Settings.NoHelp)
	viper.Set("no-title", c.Settings.NoTitle)
	viper.Set("width", c.Settings.Width)
	viper.Set("items", c.Settings.Items)
	viper.Set("volume-limit", c.Settings.VolumeLimit)
	viper.Set("volume-steps", c.Settings.VolumeSteps)
	viper.Set("no-symbols", c.Settings.NoSymbols)
	viper.Set("device-display", c.Settings.DeviceDisplay)
}
func initFlags() {
	// flag creation and validation; flag defaults are passed from validateConfig()
	flag.BoolVarP(&fullscreenFlag, "fullscreen", "f", true, "display fullscreen")
	flag.BoolVarP(&helpFlag, "no-help", "H", viper.GetBool("no-help"), "hide help text")
	flag.BoolVarP(&titleFlag, "no-title", "t", viper.GetBool("no-title"), "hide program name")
	flag.BoolVarP(&messagesFlag, "no-messages", "v", viper.GetBool("no-message"), "hide program messages")
	flag.IntVarP(&setItemsFlag, "max-items", "i", viper.GetInt("items"), "set devices per page")
	flag.IntVarP(&setWidthFlag, "max-width", "w", viper.GetInt("width"), "set width of program in terminal")
	flag.IntVarP(&maxVolumeFlag, "max-volume", "m", viper.GetInt("volume-limit"), "set maximum volume for devices")
	flag.IntVarP(&setVolumeFlag, "volume-steps", "s", viper.GetInt("volume-steps"), "set volume increments")
	flag.BoolVarP(&symbolsFlag, "no-symbols", "u", viper.GetBool("no-symbols"), "disable unicode symbols")
	flag.IntVarP(&displayFlag, "device-display", "d", viper.GetInt("device-display"), "device display level")
}
func validateFlags() {
	if setWidthFlag < minConfigWidth {
		setWidthFlag = viper.GetInt("width")
	}
	if setWidthFlag > maxConfigWidth {
		setWidthFlag = viper.GetInt("width")
	}
	if setItemsFlag < minConfigItems {
		setItemsFlag = viper.GetInt("items")
	}
	if setItemsFlag > maxConfigItems {
		setItemsFlag = viper.GetInt("items")
	}
	if maxVolumeFlag < minConfigVolume {
		maxVolumeFlag = viper.GetInt("volume-limit")
	}
	if maxVolumeFlag > maxConfigVolume {
		maxVolumeFlag = viper.GetInt("volume-limit")
	}
	if setVolumeFlag < minConfigIncrement {
		setVolumeFlag = viper.GetInt("volume-steps")
	}
	if setVolumeFlag > maxConfigIncrement {
		setVolumeFlag = viper.GetInt("volume-steps")
	}
	if displayFlag < minConfigDisplay {
		displayFlag = viper.GetInt("device-display")
	}
	if displayFlag > maxConfigDisplay {
		displayFlag = viper.GetInt("device-display")
	}
	// pass sane flag values to variables
	setAltscreen = fullscreenFlag
	setNoMessages = messagesFlag
	setNoHelp = helpFlag
	setNoTitle = titleFlag
	setWidth = setWidthFlag
	setItems = setItemsFlag
	setMaxVolume = float64(maxVolumeFlag)
	setVolume = float64(setVolumeFlag)
	setNoSymbol = symbolsFlag
	setDisplay = displayFlag
}

// load color values into program color variables
func loadConfigColors(c *Config, device, toggle *[]lipgloss.AdaptiveColor) {
	inactiveColor[0] = c.Colors.Inactive.Light
	inactiveColor[1] = c.Colors.Inactive.Dark
	activeColor[0] = c.Colors.Active.Light
	activeColor[1] = c.Colors.Active.Dark
	sinkColor[0] = c.Colors.Sink.Light
	sinkColor[1] = c.Colors.Sink.Dark
	streamColor[0] = c.Colors.Stream.Light
	streamColor[1] = c.Colors.Stream.Dark
	sourceColor[0] = c.Colors.Source.Light
	sourceColor[1] = c.Colors.Source.Dark
	outputColor[0] = c.Colors.Output.Light
	outputColor[1] = c.Colors.Output.Dark
	if errorConf { // load defaults if config file is not loaded
		c.Colors.Inactive.Light = black
		c.Colors.Inactive.Dark = White
		c.Colors.Active.Light = red
		c.Colors.Sink.Dark = White
		c.Colors.Stream.Light = green
		c.Colors.Stream.Dark = blue
		c.Colors.Source.Light = black
		c.Colors.Source.Dark = White
		c.Colors.Output.Light = green
		c.Colors.Output.Dark = blue
	}
	if setNoColor { // load defaults if config file is not loaded
		c.Colors.Inactive.Light = ""
		c.Colors.Inactive.Dark = ""
		c.Colors.Active.Light = ""
		c.Colors.Sink.Dark = ""
		c.Colors.Stream.Light = ""
		c.Colors.Stream.Dark = ""
		c.Colors.Source.Light = ""
		c.Colors.Source.Dark = ""
		c.Colors.Output.Light = ""
		c.Colors.Output.Dark = ""
	}
	*toggle = []lipgloss.AdaptiveColor{
		{Light: c.Colors.Inactive.Light, Dark: c.Colors.Inactive.Dark}, // 0 unselected item
		{Light: c.Colors.Active.Light, Dark: c.Colors.Active.Dark}}     // 1 selected item
	*device = []lipgloss.AdaptiveColor{
		{Light: c.Colors.Sink.Light, Dark: c.Colors.Sink.Dark},     // 1 sink
		{Light: c.Colors.Stream.Light, Dark: c.Colors.Stream.Dark}, // 2 stream
		{Light: c.Colors.Source.Light, Dark: c.Colors.Source.Dark}, // 3 source
		{Light: c.Colors.Output.Light, Dark: c.Colors.Output.Dark}} // 4 output
}

func loadConfigStyles(c *Config) lipgloss.Border {
	var b lipgloss.Border
	switch c.Styles.Border {
	case 1:
		b = lipgloss.HiddenBorder()
	case 2:
		b = lipgloss.RoundedBorder()
	case 3:
		b = lipgloss.ThickBorder()
	case 4:
		b = lipgloss.DoubleBorder()
	default:
		b = lipgloss.NormalBorder()
	}
	return b
}

type Config struct {
	Settings struct {
		Fullscreen    bool `mapstructure:"fullscreen"`
		NoHelp        bool `mapstructure:"nohelp"`
		NoMessage     bool `mapstructure:"nomessage"`
		NoTitle       bool `mapstructure:"notitle"`
		Width         int  `mapstructure:"width"`
		Items         int  `mapstructure:"items"`
		VolumeLimit   int  `mapstructure:"volumelimit"`
		VolumeSteps   int  `mapstructure:"volumesteps"`
		NoSymbols     bool `mapstructure:"nosymbols"`
		DeviceDisplay int  `mapstructure:"devicedisplay"`
	} `mapstructure:"settings"`
	Colors struct {
		Inactive struct {
			Light string `mapstructure:"light"`
			Dark  string `mapstructure:"dark"`
		} `mapstructure:"inactive"`
		Active struct {
			Light string `mapstructure:"light"`
			Dark  string `mapstructure:"dark"`
		} `mapstructure:"active"`
		Sink struct {
			Light string `mapstructure:"light"`
			Dark  string `mapstructure:"dark"`
		} `mapstructure:"sink"`
		Stream struct {
			Light string `mapstructure:"light"`
			Dark  string `mapstructure:"dark"`
		} `mapstructure:"stream"`
		Source struct {
			Light string `mapstructure:"light"`
			Dark  string `mapstructure:"dark"`
		} `mapstructure:"source"`
		Output struct {
			Light string `mapstructure:"light"`
			Dark  string `mapstructure:"dark"`
		} `mapstructure:"output"`
	} `mapstructure:"colors"`
	Styles struct {
		Border int `mapstructure:"border"`
	} `mapstructure:"styles"`
}
