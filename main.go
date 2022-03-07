package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	xtouch "github.com/chabad360/goxtouch"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"

	vmix "github.com/FlowingSPDG/vmix-go/http"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

var (
	wr  *writer.Writer
	wr2 *writer.Writer
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

const color = byte(xtouch.ColorWhite + xtouch.InvertLine1 + xtouch.InvertLine2)
const interval = float64(2) / 127

// This example demonstrates using github.com/chabad360/goxtouch to send and receive signals using the CTRL mode interface for the Behringer X-Touch Universal
func main() {
	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	fmt.Println(ins, outs)

	in, out := ins[0], outs[1]
	in2, out2 := ins[2], outs[4]

	must(in.Open())
	must(out.Open())
	must(in2.Open())
	must(out2.Open())

	wr = writer.New(out)
	wr.ConsolidateNotes(false)
	wr2 = writer.New(out2)
	wr2.ConsolidateNotes(false)
	rd := reader.New(reader.Each(forwardFrom), reader.NoLogger())
	rd2 := reader.New(reader.Each(forwardto), reader.NoLogger())

	go rd.ListenTo(in)
	go rd2.ListenTo(in2)

	http.Get("http://localhost:8088/API?Function=SetBusGVolume&Value=0")

	c, err := vmix.NewClient("localhost", 8088)
	must(err)

	for {
		total := 6
		if len(c.Inputs.Input) < 6 {
			total = len(c.Inputs.Input)
		}
		for i := 0; i < total; i++ {
			wr.Write(xtouch.SetLCD(xtouch.Channel1+xtouch.Channel(i), color, strconv.Itoa(i+1), c.Inputs.Input[i].Name))
			b := c.Inputs.Input[i].Balance + 1
			wr.Write(xtouch.SetEncoder(xtouch.Channel1+xtouch.Channel(i), uint8(b/interval)))
		}
		if t := 6 - total; t > 0 {
			for i := t - 1; i >= 0; i-- {
				wr.Write(xtouch.SetLCD(xtouch.Channel6-xtouch.Channel(i), byte(xtouch.ColorNone), strconv.Itoa(i+1), ""))
			}
		}
		wr.Write(xtouch.SetLCD(xtouch.Channel7, byte(xtouch.ColorYellow+48), "Preview", c.Inputs.Input[c.Preview-1].Name))
		b := c.Inputs.Input[c.Preview-1].Balance + 1
		wr.Write(xtouch.SetEncoder(xtouch.Channel7, uint8(b/interval)))
		wr.Write(xtouch.SetLCD(xtouch.Channel8, byte(xtouch.ColorGreen+48), "Active", c.Inputs.Input[c.Active-1].Name))
		b = c.Inputs.Input[c.Active-1].Balance + 1
		wr.Write(xtouch.SetEncoder(xtouch.Channel8, uint8(b/interval)))
		time.Sleep(time.Millisecond * 20)
		c, err = vmix.NewClient("localhost", 8088)
		must(err)
	}
}

func forwardto(p *reader.Position, m midi.Message) {
	wr.Write(m)
}

func forwardFrom(p *reader.Position, m midi.Message) {
	wr2.Write(m)
}
