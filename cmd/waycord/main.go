package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/x/explorer"

	"github.com/angyedz/Waycord/internal/core/attachments"
	"github.com/angyedz/Waycord/internal/ota"
	"github.com/angyedz/Waycord/internal/storage"
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
	chooser := explorer.NewExplorer(w)
	view.ConfigureAttachments(func() (*attachments.Attachment, error) {
		file, err := chooser.ChooseFile()
		if errors.Is(err, explorer.ErrUserDecline) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		defer file.Close()
		name := "Attachment"
		if named, ok := file.(interface{ Name() string }); ok {
			name = named.Name()
		}
		return attachments.Read(name, file)
	})
	if configDir, err := os.UserConfigDir(); err != nil {
		log.Print("Waycord update preferences: ", err)
	} else {
		store := storage.UpdatePreferencesStore{Root: filepath.Join(configDir, "waycord")}
		preferences, loadErr := store.Load()
		if loadErr != nil {
			log.Print("Waycord update preferences: ", loadErr)
			preferences = storage.UpdatePreferences{Channel: "stable"}
		}
		view.ConfigureUpdatePreferences(preferences.Channel, preferences.AutoCheck, func(channel string, autoCheck bool) error {
			return store.Save(storage.UpdatePreferences{Channel: channel, AutoCheck: autoCheck})
		})
	}
	updateCtx, stopUpdateChecks := context.WithCancel(context.Background())
	defer stopUpdateChecks()
	view.StartUpdateChecks(updateCtx, w.Invalidate)
	var ops op.Ops
	healthReported := false
	var nextHealthPulse time.Time
	var healthObservationEnds time.Time
	for {
		event := w.Event()
		chooser.ListenEvents(event)
		switch event := event.(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)
			view.Layout(gtx)
			challenge := ota.HealthChallengeEnabled()
			if challenge && !healthReported {
				gtx.Execute(op.InvalidateCmd{At: time.Now().Add(ota.HealthPulseInterval)})
			} else if challenge && time.Now().Before(healthObservationEnds) {
				now := time.Now()
				if !now.Before(nextHealthPulse) {
					if err := ota.PulseHealthy(); err != nil {
						log.Print("startup heartbeat: ", err)
					}
					nextHealthPulse = now.Add(ota.HealthPulseInterval)
				}
				if nextHealthPulse.IsZero() {
					nextHealthPulse = now.Add(ota.HealthPulseInterval)
				}
				next := nextHealthPulse
				if healthObservationEnds.Before(next) {
					next = healthObservationEnds
				}
				gtx.Execute(op.InvalidateCmd{At: next})
			}
			event.Frame(gtx.Ops)
			if !healthReported {
				healthReported = true
				if err := ota.MarkHealthy(); err != nil {
					log.Print("startup handshake: ", err)
				}
				nextHealthPulse = time.Now().Add(ota.HealthPulseInterval)
				healthObservationEnds = time.Now().Add(ota.HealthObservationWindow)
			}
		}
	}
}
