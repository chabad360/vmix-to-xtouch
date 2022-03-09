package main

import (
	_ "embed"
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	driver "gitlab.com/gomidi/rtmididrv"
)

func gui() {
	w := a.NewWindow("VMix to X-Touch")
	//w.SetIcon(logoResource)

	infoLabel := widget.NewRichTextWithText("Forwarding Off")
	infoLabel.Wrapping = fyne.TextWrapOff

	vmixPort := widget.NewEntry()
	vmixPort.SetText(VMixPort)
	vmixPort.Validator = validation.NewRegexp(`^[0-9]*$`, "not a valid port")

	vmixAddr := widget.NewEntry()
	vmixAddr.SetText(VMixAddr)
	vmixAddr.Validator = validation.NewRegexp(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, "not a valid IP address")

	ins, outs := insAndOuts(drv)

	xtouchIn := widget.NewSelectEntry(ins)
	xtouchIn.Validator = listValidator(ins)
	xtouchIn.Text = a.Preferences().String("xtouchin")

	xtouchOut := widget.NewSelectEntry(outs)
	xtouchOut.Validator = listValidator(outs)
	xtouchOut.Text = a.Preferences().String("xtouchout")

	activator := widget.NewSelectEntry(ins)
	activator.Validator = listValidator(ins)
	activator.Text = a.Preferences().String("activator")

	shortcut := widget.NewSelectEntry(outs)
	shortcut.Validator = listValidator(outs)
	shortcut.Text = a.Preferences().String("shortcut")

	refreshButton := widget.NewButton("Refresh List", func() {
		ins, outs = insAndOuts(drv)
		xtouchIn.SetOptions(ins)
		xtouchIn.Validator = listValidator(ins)
		xtouchOut.SetOptions(outs)
		xtouchOut.Validator = listValidator(outs)

		activator.SetOptions(ins)
		activator.Validator = listValidator(ins)
		shortcut.SetOptions(outs)
		shortcut.Validator = listValidator(outs)
	})

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "VMix API Address", Widget: vmixAddr, HintText: "VMix API Address (this should be 127.0.0.1)"},
			{Text: "VMix API Port", Widget: vmixPort, HintText: "VMix API Port (usually 8088)"},
			{Text: "VMix Activator Loop Port", Widget: activator, HintText: "This should be a LoopMIDI port"},
			{Text: "VMix Shortcut Loop Port", Widget: shortcut, HintText: "This should be a LoopMIDI port"},
			{Text: "X-Touch Input Port", Widget: xtouchIn, HintText: "This should be called X-Touch 0"},
			{Text: "X-Touch Output Port", Widget: xtouchOut, HintText: "This should be called X-Touch 0"},
			{Text: "Refresh MIDI list", Widget: refreshButton, HintText: "Refresh the available MIDI ports"},
		},
		SubmitText: "Start Forwarding",
		CancelText: "Stop Forwarding",
	}

	form.OnCancel = nil
	form.OnSubmit = func() {
		infoLabel.ParseMarkdown("Starting Forwarding")
		running = false
		wg.Wait()
		VMixPort = vmixPort.Text
		VMixAddr = vmixAddr.Text

		i, err := drv.Ins()
		if err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Forwarding Errored")
			return
		}
		o, err := drv.Outs()
		if err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Forwarding Errored")
			return
		}

		if f, ok := getInList(activator.Text, ins); ok {
			Activator = i[f]
			a.Preferences().SetString("activator", activator.Text)
		} else {
			infoLabel.ParseMarkdown("Forwarding Errored")
			return
		}
		if f, ok := getInList(shortcut.Text, outs); ok {
			Shortcut = o[f]
			a.Preferences().SetString("shortcut", shortcut.Text)
		} else {
			infoLabel.ParseMarkdown("Forwarding Errored")
			return
		}

		if f, ok := getInList(xtouchIn.Text, ins); ok {
			XTouchIn = i[f]
			a.Preferences().SetString("xtouchin", xtouchIn.Text)
		} else {
			infoLabel.ParseMarkdown("Forwarding Errored")
			return
		}
		if f, ok := getInList(xtouchOut.Text, outs); ok {
			XTouchOut = o[f]
			a.Preferences().SetString("xtouchout", xtouchOut.Text)
		} else {
			infoLabel.ParseMarkdown("Forwarding Errored")
			return
		}

		running = true

		if err = startMidi(); err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Forwarding Errored")
			return
		}

		infoLabel.ParseMarkdown(fmt.Sprintf("Forwarding messages."))
		form.SubmitText = "Update Server"

		form.OnCancel = func() {
			running = false
			wg.Wait()

			form.OnCancel = nil
			form.SubmitText = "Start Forwarding"

			form.Refresh()
			runtime.GC()
		}

		form.Refresh()
		runtime.GC()
	}

	w.SetContent(container.NewGridWithRows(2, form, infoLabel))
	w.ShowAndRun()
}

func insAndOuts(d *driver.Driver) (i []string, o []string) {
	ins, _ := d.Ins()
	outs, _ := d.Outs()
	i = make([]string, len(ins))
	for f := 0; f < len(ins); f++ {
		i[f] = ins[f].String()
	}
	o = make([]string, len(outs))
	for f := 0; f < len(outs); f++ {
		o[f] = outs[f].String()
	}
	return
}

func getInList(item string, items []string) (int, bool) {
	for i := 0; i < len(items); i++ {
		if items[i] == item {
			return i, true
		}
	}
	return -1, false
}

func listValidator(items []string) func(string) error {
	return func(s string) error {
		if _, ok := getInList(s, items); !ok {
			return fmt.Errorf("invalid entry")
		}
		return nil
	}
}
