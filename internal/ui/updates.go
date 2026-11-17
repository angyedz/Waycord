package ui

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"runtime"
	"time"

	"gioui.org/layout"

	"github.com/angyedz/Waycord/internal/ota"
	"github.com/angyedz/Waycord/internal/platform"
	"github.com/angyedz/Waycord/internal/version"
)

const automaticUpdateInterval = 12 * time.Hour

type updateCheckResult struct {
	channel string
	update  ota.Update
	err     error
}

// ConfigureUpdatePreferences binds the screen to persisted user preferences.
func (a *App) ConfigureUpdatePreferences(channel string, autoCheck bool, save func(string, bool) error) {
	if channel != "stable" && channel != "beta" && channel != "nightly" {
		channel = "stable"
	}
	a.updateChannel, a.updateAutoCheck, a.saveUpdatePreferences = channel, autoCheck, save
}

func (a *App) setUpdatePreferences(channel string, autoCheck bool) {
	if channel != "stable" && channel != "beta" && channel != "nightly" {
		return
	}
	if a.saveUpdatePreferences != nil {
		if err := a.saveUpdatePreferences(channel, autoCheck); err != nil {
			a.updateStatus = "Couldn't save update preferences: " + err.Error()
			return
		}
	}
	changed := channel != a.updateChannel
	a.updateChannel, a.updateAutoCheck = channel, autoCheck
	if changed {
		a.availableUpdate = nil
		a.updateStatus = ""
	}
}

// StartUpdateChecks starts a quiet periodic scheduler. It never makes a request
// unless automatic checks are enabled and a trusted release key is configured.
func (a *App) StartUpdateChecks(ctx context.Context, invalidate func()) {
	a.updateCtx = ctx
	a.updateInvalidate = invalidate
	if a.updateAutoCheck && signingKeyConfigured() {
		a.checkForUpdates()
	}
	go func() {
		ticker := time.NewTicker(automaticUpdateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case a.updateSignals <- struct{}{}:
				default:
				}
				if invalidate != nil {
					invalidate()
				}
			}
		}
	}()
}

func (a *App) processUpdateChecks() {
	for {
		select {
		case result := <-a.updateResult:
			a.updateChecking = false
			if result.channel != a.updateChannel {
				continue
			}
			if result.err != nil {
				a.availableUpdate = nil
				a.updateStatus = result.err.Error()
			} else if result.update.Available {
				a.availableUpdate = &result.update
				a.updateStatus = "A signed update is available."
			} else {
				a.availableUpdate = nil
				a.updateStatus = "You're up to date."
			}
		case <-a.updateSignals:
			if a.updateAutoCheck && !a.updateChecking {
				a.checkForUpdates()
			}
		default:
			return
		}
	}
}

func (a *App) checkForUpdates() {
	if a.updateChecking {
		return
	}
	key, err := hex.DecodeString(version.SigningPublicKeyHex)
	if err != nil || len(key) != ed25519.PublicKeySize {
		a.availableUpdate = nil
		a.updateStatus = "Signed updates aren't enabled for this build yet."
		return
	}
	a.updateChecking = true
	a.updateStatus = "Checking for updates…"
	ctx, cancel := context.WithTimeout(a.updateCtx, 3*time.Minute)
	go func(channel string, publicKey ed25519.PublicKey) {
		defer cancel()
		update, checkErr := a.updateChecker.Check(ctx, channel, runtime.GOOS, runtime.GOARCH, version.Version, publicKey)
		a.updateResult <- updateCheckResult{channel: channel, update: update, err: checkErr}
		if a.updateInvalidate != nil {
			a.updateInvalidate()
		}
	}(a.updateChannel, ed25519.PublicKey(key))
}

func signingKeyConfigured() bool {
	key, err := hex.DecodeString(version.SigningPublicKeyHex)
	return err == nil && len(key) == ed25519.PublicKeySize
}

func (a *App) updates(gtx layout.Context) layout.Dimensions {
	a.updateList.Axis = layout.Vertical
	return a.updateList.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		return a.updatesContent(gtx)
	})
}

func (a *App) updatesContent(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 16, Bottom: 20, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(a.text("Updates", 24, a.palette.Text)),
			vspace(4),
			layout.Rigid(a.text("Choose how Waycord checks for new versions.", 14, a.palette.Muted)),
			vspace(18),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 14, padding(layout.Inset{Left: 18, Right: 18, Top: 15, Bottom: 15}, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("Installed version", 12, a.palette.Subtle)),
						vspace(5),
						layout.Rigid(a.text(fmt.Sprintf("Waycord  %s", version.Version), 17, a.palette.Text)),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.updateStatus == "" {
								if !signingKeyConfigured() {
									return a.text("Signed update checks aren't enabled for this build.", 13, a.palette.Muted)(gtx)
								}
								return a.text("Updates are verified with a trusted release signature.", 13, a.palette.Muted)(gtx)
							}
							color := a.palette.Muted
							if a.availableUpdate != nil {
								color = a.palette.Positive
							}
							return a.text(a.updateStatus, 13, color)(gtx)
						}),
					)
				}))
			}),
			vspace(16),
			layout.Rigid(a.text("Release channel", 15, a.palette.Text)),
			vspace(8),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				channels := [...]string{"stable", "beta", "nightly"}
				children := make([]layout.FlexChild, 0, len(channels)*2)
				for i, channel := range channels {
					i, channel := i, channel
					children = append(children,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.pill(gtx, &a.updateChannelButtons[i], channel, a.updateChannel == channel)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if i == len(channels)-1 {
								return layout.Dimensions{}
							}
							return layout.Spacer{Width: 7}.Layout(gtx)
						}),
					)
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
			}),
			vspace(16),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := "Automatic checks  ·  Off"
				if a.updateAutoCheck {
					label = "Automatic checks  ·  On"
				}
				return a.pill(gtx, &a.updateAutoButton, label, a.updateAutoCheck)
			}),
			vspace(12),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := "Check for updates"
				if a.updateChecking {
					label = "Checking…"
				}
				return a.pill(gtx, &a.updateCheckButton, label, false)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.availableUpdate == nil {
					return layout.Dimensions{}
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					vspace(18),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return rounded(gtx, a.palette.Surface, 14, inset(16, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(a.text("Version "+a.availableUpdate.Payload.Version, 17, a.palette.Text)),
								vspace(8),
								layout.Rigid(a.text(a.availableUpdate.Notes, 14, a.palette.Muted)),
							)
						}))
					}),
					vspace(8),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.pill(gtx, &a.updateOpenButton, "Release notes", false)
					}),
				)
			}),
		)
	})
}

func openReleasePage(address string) error { return platform.OpenURL(address) }
