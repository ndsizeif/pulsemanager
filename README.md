## Pulsemanager

Control your [pulseaudio](https://www.freedesktop.org/wiki/Software/PulseAudio/) devices with a colorful terminal user interface via [bubbletea](https://github.com/charmbracelet/bubbletea). 

<p>
  <img src="https://github.com/ndsizeif/pulsemanager/blob/assets/assets/demo.gif?"/>
	<br>
</p>  

### Installation
Download the main branch of the project without any assets.
```
git clone --depth=1 https://github.com/ndsizeif/pulsemanager
```
To try it out simply type `go run .` in the project directory.
To install run `go build` and place the pulsemanager binary in your PATH.
The system must be using and running a pulseaudio sound server and have pactl
installed.

### Usage

Pulsemanager can be used to view audio devices running on a pulseaudio server.
It can issue basic pactl commands to control device parameters, such as volume.

#### Flags

```
  -f, --fullscreen         display fullscreen (default true)
  -i, --max-items int      set devices per page (default 5)
  -m, --max-volume int     set maximum volume for devices (default 180)
  -w, --max-width int      set width of program in terminal (default 100)
  -H, --no-help            hide help text
  -v, --no-messages        hide program messages
  -t, --no-title           hide program name
  -s, --volume-steps int   set volume increments (default 5)
```
#### Implemented Commands

```
 * set-default-sink
 * set-default-souce
 * set-sink-mute
 * set-sink-input-mute
 * set-source-mute 
 * set-source-output-mute 
 * set-sink-volume
 * set-sink-input-volume
 * set-source-volume
 * set-source-output-volume
 * move-sink-input
 * move-source-output
 * load-module module-loopback
 * unload-module module-loopback
```

#### Configuration

A few pulsemanager settings can be configured using `config.yaml`. This file can
be saved in the local directory from which the application is launched.
Alternatively, create `$HOME/.config/pulsemanager/` and save it there.

Change the values in the [example configuration](example/config.yaml) to alter
the [color
scheme](https://github.com/ndsizeif/pulsemanager/blob/assets/assets/colors.gif) and
appearance. There are also a few basic preferences.

The maximum terminal width can be adjusted. However, if the terminal is re-sized
and is too small (less than 45 columns or 8 rows), the program will exit.

#### Key Bindings

| key     | function          | Description                                   |   |
|---------|-------------------|-----------------------------------------------|---|
| j       | cursor down       | go to next device (down)                      |   |
| k       | cursor up         | go to previous device (up)                    |   |
| n       | next page         | show next page of devices (page down)         |   |
| p/N     | previous page     | show previous page of devices (page up)       |   |
| h/J     | volume down       | decrease volume by increment (left)           |   |
| l/K     | volume up         | increase volume by increment (right)          |   |
| g       | first entry       | go to first device (home)                     |   |
| G       | last entry        | go to last device (end)                       |   |
| m/Space | toggle mute       | un(mute) a device or stream                   |   |
| v       | show messages     | messages can be turned off in config          |   |
| c       | change channel    | cycle through and control individual channels |   |
| s       | select device     | toggle a device as target for a command       | * |
| Escape  | cancel selection  | deselect device or channel                    |   |
| x       | terminate stream  | parent process may re-spawn the stream        |   |
| X       | unload loopback   | all loopback streams will be terminated       |   |
| f       | toggle fullscreen | program launches fullscreen by default        |   |
| t       | change display    | display less or more device information       |   |
| Enter   | perform action    | command depends on type of device selected    | * |
| 1-0     | set device volume | set volume of all channels from 10% to 100%   |   |
| -       | decrease latency  | value is used when loading loopback module    | * |
| =/+     | increase latency  | value is used when loading loopback module    |   |
| q       | exit              | terminate program (ctrl-c)                    |   |

Select Device
- The currently toggled device will be displayed in upper left.

Perform Action
- Pressing enter on a sink/source will make it the [default sink/source](https://github.com/ndsizeif/pulsemanager/blob/assets/assets/inlinedemo.gif) if no device is toggled. 
- If a sink/source is toggled, pressing enter on that device will (un)suspend it.
- If a sink is toggled, pressing enter on a source will create a [loopback stream](https://github.com/ndsizeif/pulsemanager/blob/assets/assets/loopbackdemo.gif?).
- If a sink is toggled, pressing enter on a stream will move the stream to it.
- If a source is toggled, pressing enter on an output will move the output to it.

Latency
- The adjustable latency range for loopback module is 10 - 500 milliseconds.


### Issues

Pactl 16.x json format doesn't handle [non-ascii characters](https://gitlab.freedesktop.org/pulseaudio/pulseaudio/-/issues/1310).
This requires media name parsing from the pactl text format. The program may
briefly report inaccurate data.

Moving an output to another source needs testing.

### Contributing

This project was primarily to test out golang and the bubbletea library. But any
reports of blatant bugs are appreciated. The next version will eschew pactl,
while offering more commands and visualization. Request for additional features
will likely end up there.
