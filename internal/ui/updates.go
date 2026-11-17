package ui

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"image"
	"runtime"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

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

type updateInstallProgress struct {
	downloaded int64
	total      int64
}

type updateInstallResult struct{ err error }

// ConfigureUpdateInstallation enables in-app staging only for a verified,
// user-managed install. Package-managed installs leave both callbacks nil.
func (a *App) ConfigureUpdateInstallation(
	install func(context.Context, ota.Update, string, ed25519.PublicKey, ota.ProgressFunc) error,
	restart func() error,
) {
	if install == nil || restart == nil {
		a.updateInstaller, a.restartUpdatedApp = nil, nil
		return
	}
	a.updateInstaller, a.restartUpdatedApp = install, restart
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
			a.updateStatus = "Couldn't save update preferences."
			return
		}
	}
	changed := channel != a.updateChannel
	a.updateChannel, a.updateAutoCheck = channel, autoCheck
	if changed {
		if !a.updateStaged {
			a.availableUpdate = nil
			a.updateStatus = ""
		}
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
				a.updateStatus = "Couldn't check for updates. Try again later."
			} else if result.update.Available {
				a.availableUpdate = &result.update
				a.updateStatus = "An update is available."
			} else {
				a.availableUpdate = nil
				a.updateStatus = "You're up to date."
			}
		case progress := <-a.updateProgress:
			if a.updateInstalling {
				a.updateDownloaded, a.updateTotal = progress.downloaded, progress.total
			}
		case result := <-a.updateInstallResult:
			a.drainUpdateProgress()
			a.updateInstalling = false
			if result.err != nil {
				a.updateStaged = false
				a.updateStatus = "Couldn't download the update. Please try again."
			} else {
				a.updateStaged = true
				a.updateStatus = "Update ready to install. Restart Waycord when you're ready."
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

func (a *App) drainUpdateProgress() {
	for {
		select {
		case progress := <-a.updateProgress:
			a.updateDownloaded, a.updateTotal = progress.downloaded, progress.total
		default:
			return
		}
	}
}

func (a *App) installAvailableUpdate() {
	if a.updateInstalling || a.updateStaged || a.availableUpdate == nil || a.updateInstaller == nil {
		return
	}
	key, err := ed25519PublicKey()
	if err != nil {
		a.updateStatus = "Updates are unavailable for this installation."
		return
	}
	a.updateInstalling = true
	a.updateDownloaded, a.updateTotal = 0, a.availableUpdate.Payload.Size
	a.updateStatus = "Downloading update…"
	ctx, cancel := context.WithTimeout(a.updateCtx, 5*time.Minute)
	update := *a.availableUpdate
	channel := a.updateChannel
	publicKey := append(ed25519.PublicKey(nil), key...)
	go func() {
		defer cancel()
		progress := func(downloaded, total int64) {
			latest := updateInstallProgress{downloaded: downloaded, total: total}
			select {
			case a.updateProgress <- latest:
			default:
				select {
				case <-a.updateProgress:
				default:
				}
				select {
				case a.updateProgress <- latest:
				default:
				}
			}
			if a.updateInvalidate != nil {
				a.updateInvalidate()
			}
		}
		installErr := a.updateInstaller(ctx, update, channel, publicKey, progress)
		a.updateInstallResult <- updateInstallResult{err: installErr}
		if a.updateInvalidate != nil {
			a.updateInvalidate()
		}
	}()
}

func (a *App) checkForUpdates() {
	if a.updateChecking {
		return
	}
	key, err := ed25519PublicKey()
	if err != nil {
		a.availableUpdate = nil
		a.updateStatus = "Updates are unavailable for this installation."
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
	_, err := ed25519PublicKey()
	return err == nil
}

func ed25519PublicKey() (ed25519.PublicKey, error) {
	key, err := hex.DecodeString(version.SigningPublicKeyHex)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid signing public key")
	}
	return ed25519.PublicKey(key), nil
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
									return a.text("Update checks are unavailable for this installation.", 13, a.palette.Muted)(gtx)
								}
								return a.text("Waycord will let you know when an update is ready.", 13, a.palette.Muted)(gtx)
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
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if !a.updateInstalling {
										return layout.Dimensions{}
									}
									return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
										vspace(12),
										layout.Rigid(a.updateProgressBar),
										vspace(5),
										layout.Rigid(a.text(a.updateProgressLabel(), 12, a.palette.Muted)),
									)
								}),
							)
						}))
					}),
					vspace(8),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.pill(gtx, &a.updateOpenButton, "Release notes", false)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if a.updateInstaller == nil {
							return layout.Inset{Left: 8}.Layout(gtx, a.text("Use your usual installation method to update Waycord.", 13, a.palette.Muted))
						}
						if a.updateInstalling {
							return layout.Dimensions{}
						}
						return layout.Spacer{Width: 8}.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if a.updateInstaller == nil {
							return layout.Dimensions{}
						}
						if a.updateInstalling {
							return layout.Dimensions{}
						}
						label, button := "Install update", &a.updateInstallButton
						if a.updateStaged {
							label, button = "Restart to update", &a.updateRestartButton
						}
						return a.pill(gtx, button, label, a.updateStaged)
					}),
				)
			}),
		)
	})
}

func (a *App) updateProgressLabel() string {
	if a.updateTotal <= 0 {
		return "Downloading update…"
	}
	pct := a.updateDownloaded * 100 / a.updateTotal
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("Downloading update… %d%%", pct)
}

func (a *App) updateProgressBar(gtx layout.Context) layout.Dimensions {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(4))
	if size.X <= 0 {
		return layout.Dimensions{}
	}
	track := image.Rectangle{Max: size}
	paint.FillShape(gtx.Ops, a.palette.Surface2, clip.UniformRRect(track, gtx.Dp(2)).Op(gtx.Ops))
	if a.updateTotal > 0 && a.updateDownloaded > 0 {
		fraction := float64(a.updateDownloaded) / float64(a.updateTotal)
		if fraction > 1 {
			fraction = 1
		}
		fill := image.Rect(0, 0, int(float64(size.X)*fraction), size.Y)
		if fill.Dx() > 0 {
			paint.FillShape(gtx.Ops, a.palette.Accent, clip.UniformRRect(fill, gtx.Dp(2)).Op(gtx.Ops))
		}
	}
	return layout.Dimensions{Size: size}
}

func openReleasePage(address string) error { return platform.OpenURL(address) }
