package main

import (
	"fyne.io/fyne/v2/app"
	vmix "github.com/FlowingSPDG/vmix-go/http"
	xtouch "github.com/chabad360/goxtouch"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"strconv"
	"sync"
	"time"

	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

var (
	wr  *writer.Writer
	wr2 *writer.Writer

	a   = app.NewWithID("me.chabad360.vmix-to-xtouch")
	drv *driver.Driver

	VMixPort = "8088"
	VMixAddr = "127.0.0.1"

	Activator midi.In
	Shortcut  midi.Out
	XTouchIn  midi.In
	XTouchOut midi.Out

	running = false
	wg      = new(sync.WaitGroup)
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

const color = byte(xtouch.ColorWhite + xtouch.InvertLine1 + xtouch.InvertLine2)
const interval = float64(2) / 127

func main() {
	drv, _ = driver.New()
	defer drv.Close()
	gui()
}

func startMidi() error {
	if err := Activator.Open(); err != nil {
		return err
	}
	if err := Shortcut.Open(); err != nil {
		return err
	}
	if err := XTouchOut.Open(); err != nil {
		return err
	}
	if err := XTouchIn.Open(); err != nil {
		return err
	}

	wr = writer.New(XTouchOut)
	wr.ConsolidateNotes(false)
	wr2 = writer.New(Shortcut)
	wr2.ConsolidateNotes(false)
	rd := reader.New(reader.Each(forwardFrom), reader.NoLogger())
	rd2 := reader.New(reader.Each(forwardto), reader.NoLogger())

	go func() {
		wg.Add(1)
		defer wg.Done()
		rd.ListenTo(XTouchIn)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		rd2.ListenTo(Activator)
	}()

	//http.Get("http://localhost:8088/API?Function=SetBusGVolume&Value=0")

	p, _ := strconv.Atoi(VMixPort)
	c, err := vmix.NewClient(VMixAddr, p)
	if err != nil {
		return err
	}

	go func() {
		wg.Add(1)
		defer wg.Done()
		for running {
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
			p, _ = strconv.Atoi(VMixPort)
			c, err = vmix.NewClient(VMixAddr, p)
			must(err)
		}
		drv.Close()
		drv, _ = driver.New()
	}()

	return nil
}

func forwardto(p *reader.Position, m midi.Message) {
	wr.Write(m)
}

func forwardFrom(p *reader.Position, m midi.Message) {
	wr2.Write(m)
}
