module vmix-to-xtouch

go 1.17

replace github.com/chabad360/goxtouch => ../goxtouch

require (
	github.com/FlowingSPDG/vmix-go v0.2.3
	github.com/chabad360/goxtouch v0.0.0-00010101000000-000000000000
	gitlab.com/gomidi/midi v1.23.7
	gitlab.com/gomidi/rtmididrv v0.14.0
)
