// /////////////////////////////////////////////////////////////////////////////
// LAUNCH APPLICATION
// /////////////////////////////////////////////////////////////////////////////
package main

import (
	"fmt"                                    // format and print text
	tea "github.com/charmbracelet/bubbletea" // main cli application library
	flag "github.com/cornfeedhobo/pflag"     // command line flag parsing
)

func main() {
	setNoColor = haveNoColor()  // let program know if "NO_COLOR" is set
	setDefaults(initDefaults()) // set the default values for program
	var config Config           // make a config struct for population
	readConfig(&config)         // read config file
	validateConfig(&config)     // validate config file and load its settings
	loadConfigColors(&config, &deviceColor, &toggleColor)
	setBorder = loadConfigStyles(&config)
	initFlags()
	flag.Parse()
	validateFlags()
	// check for dependencies
	if !haveProgram("pactl") {
		fmt.Println(errorOut(errorMsg1))
		return
	}
	istty = isConsole()
	if istty || setNoSymbol {
		disableSymbols()
	}
	// launch bubbletea fullscreen or inline
	var p *tea.Program
	if setAltscreen {
		p = tea.NewProgram(setupModel(), tea.WithAltScreen())
	} else {
		p = tea.NewProgram(setupModel())
	}
	if err := p.Start(); err != nil {
		fmt.Printf("Error initializing program: %v", err)
	}
}
