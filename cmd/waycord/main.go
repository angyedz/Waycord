package main

import (
	"log"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/angyedz/Waycord/internal/ota"
	"github.com/angyedz/Waycord/internal/ui"
)

func main() {
	go func() {
		if err := run(); err != nil {
			log.Print("Waycord: ", err)
		}
	}()
	app.Main()
}

func run() error {
	w := new(app.Window)
	w.Option(app.Title("Waycord — Preview"), app.Size(unit.Dp(1100), unit.Dp(740)), app.MinSize(unit.Dp(800), unit.Dp(540)))
	view := ui.New()
	var ops op.Ops
	healthReported := false
	for {
		switch event := w.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)
			view.Layout(gtx)
			event.Frame(gtx.Ops)
			if !healthReported {
				healthReported = true
				if err := ota.MarkHealthy(); err != nil {
					log.Print("startup handshake: ", err)
				}
			}
		}
	}
}
