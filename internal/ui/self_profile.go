package ui

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var selfPresenceNames = [...]string{"Online", "Idle", "Do not disturb", "Invisible"}

func (a *App) processSelfProfile(gtx layout.Context) {
	if a.selfProfileButton.Clicked(gtx) {
		a.selfProfileOpen = !a.selfProfileOpen
		a.selfProfileEditing = false
		a.selfProfileError = ""
		if a.selfProfileOpen {
			a.contextTarget = -1
		}
	}
	if !a.selfProfileOpen && a.selfProfileSlide.value <= 0 {
		return
	}
	if a.selfProfileBackdrop.Clicked(gtx) {
		a.selfProfileOpen = false
		a.selfProfileEditing = false
	}
	if a.selfProfileEditing {
		for {
			if _, ok := a.selfProfileNameEditor.Update(gtx); !ok {
				break
			}
		}
		for {
			if _, ok := a.selfProfileStatusEditor.Update(gtx); !ok {
				break
			}
		}
		if a.selfProfileCancelButton.Clicked(gtx) {
			a.selfProfileEditing = false
		}
		if a.selfProfileSaveButton.Clicked(gtx) {
			a.saveSelfProfile()
		}
	} else {
		if a.selfProfileEditButton.Clicked(gtx) {
			a.selfProfileEditing = true
			a.selfProfileNameEditor.SetText(a.selfProfileName)
			a.selfProfileStatusEditor.SetText(a.selfProfileStatus)
			gtx.Execute(key.FocusCmd{Tag: &a.selfProfileNameEditor})
		}
		for i := range a.selfStatusButtons {
			if a.selfStatusButtons[i].Clicked(gtx) {
				a.setLocalPresence(i)
			}
		}
	}
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if event, ok := e.(key.Event); ok && event.State == key.Press {
			a.selfProfileOpen = false
			a.selfProfileEditing = false
		}
	}
}

func (a *App) ConfigureLocalProfile(name, status string, presence int, save func(string, string, int) error) {
	if strings.TrimSpace(name) == "" {
		name = "angyedz"
	}
	if presence < 0 || presence >= len(selfPresenceNames) {
		presence = 3
	}
	a.selfProfileName, a.selfProfileStatus, a.selfPresence, a.saveLocalProfile = name, status, presence, save
	a.selfProfileNameEditor.SetText(a.selfProfileName)
	a.selfProfileStatusEditor.SetText(a.selfProfileStatus)
}

func (a *App) saveSelfProfile() {
	name := strings.TrimSpace(a.selfProfileNameEditor.Text())
	if name == "" {
		name = a.selfProfileName
	}
	status := strings.TrimSpace(a.selfProfileStatusEditor.Text())
	if a.saveLocalProfile != nil {
		if err := a.saveLocalProfile(name, status, a.selfPresence); err != nil {
			a.selfProfileError = "Couldn't save profile changes. Try again."
			return
		}
	}
	a.selfProfileName, a.selfProfileStatus = name, status
	a.selfProfileError = ""
	a.selfProfileNameEditor.SetText(a.selfProfileName)
	a.selfProfileStatusEditor.SetText(a.selfProfileStatus)
	a.selfProfileEditing = false
}

func (a *App) setLocalPresence(presence int) {
	if presence < 0 || presence >= len(selfPresenceNames) {
		return
	}
	if a.saveLocalProfile != nil {
		if err := a.saveLocalProfile(a.selfProfileName, a.selfProfileStatus, presence); err != nil {
			a.selfProfileError = "Couldn't save profile changes. Try again."
			return
		}
	}
	a.selfPresence = presence
	a.selfProfileError = ""
	a.selfProfileOpen = false
}

func (a *App) selfProfileOverlay(gtx layout.Context, size image.Point, progress float32) {
	backdrop := gtx
	backdrop.Constraints = layout.Exact(size)
	a.selfProfileBackdrop.Layout(backdrop, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{A: uint8(24 * progress)}, clip.Rect{Max: size}.Op())
		return layout.Dimensions{Size: size}
	})

	cardContext := gtx
	cardContext.Constraints = layout.Constraints{Max: image.Pt(min(gtx.Dp(288), size.X), max(0, size.Y-gtx.Dp(76)))}
	record := op.Record(gtx.Ops)
	card := a.selfProfileCard(cardContext)
	content := record.Stop()
	top := max(0, size.Y-gtx.Dp(76)-card.Size.Y)
	position := image.Pt(gtx.Dp(8), top+int(float32(gtx.Dp(12))*(1-progress)))
	layer := op.Record(gtx.Ops)
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	offset := op.Offset(position).Push(gtx.Ops)
	content.Add(gtx.Ops)
	offset.Pop()
	area.Pop()
	op.Defer(gtx.Ops, layer.Stop())
}

func (a *App) selfProfileCard(gtx layout.Context) layout.Dimensions {
	return fixedWidth(288, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Sidebar, 16, padding(layout.Inset{Top: 14, Bottom: 12, Left: 12, Right: 12}, func(gtx layout.Context) layout.Dimensions {
			if a.selfProfileEditing {
				return a.selfProfileEditor(gtx)
			}
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return fixedWidth(56, fixedHeight(56, func(gtx layout.Context) layout.Dimensions {
								return rounded(gtx, a.palette.AccentDim, 28, func(gtx layout.Context) layout.Dimensions {
									return layout.Center.Layout(gtx, a.text(firstLetter(a.selfProfileName), 22, a.palette.Text))
								})
							}))(gtx)
						}),
						hspace(12),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(a.text(a.selfProfileName, 19, a.palette.Text)),
								layout.Rigid(a.text("@"+a.selfProfileUsername, 12, a.palette.Subtle)),
							)
						}),
					)
				}),
			}
			if a.selfProfileStatus != "" {
				children = append(children, vspace(10), layout.Rigid(a.text(a.selfProfileStatus, 13, a.palette.Muted)))
			}
			if a.voice.Connected {
				children = append(children, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.Surface, 10, inset(10, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(a.text("In voice", 12, a.palette.Positive)),
							layout.Rigid(a.text(a.voiceChannel+" · "+a.voiceServer, 13, a.palette.Text)),
						)
					}))
				}))
			}
			children = append(children,
				vspace(14),
				layout.Rigid(a.text("SET STATUS", 11, a.palette.Subtle)),
				vspace(5),
			)
			for i, status := range selfPresenceNames {
				i, status := i, status
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.clickLabel(gtx, &a.selfStatusButtons[i], status, a.selfPresence == i)
				}))
			}
			children = append(children,
				vspace(10),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.pill(gtx, &a.selfProfileEditButton, "Edit profile", false)
				}),
			)
			if a.selfProfileError != "" {
				children = append(children, vspace(8), layout.Rigid(a.text(a.selfProfileError, 12, a.palette.Warning)))
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		}))
	})(gtx)
}

func (a *App) selfProfileEditor(gtx layout.Context) layout.Dimensions {
	field := func(editor *widget.Editor, hint string) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			input := material.Editor(a.theme, editor, hint)
			input.TextSize = 14
			input.Color = a.palette.Text
			return rounded(gtx, a.palette.Surface, 9, padding(layout.Inset{Left: 10, Right: 10, Top: 8, Bottom: 8}, input.Layout))
		}
	}
	children := []layout.FlexChild{
		layout.Rigid(a.text("Edit profile", 19, a.palette.Text)),
		vspace(12),
		layout.Rigid(a.text("Display name", 12, a.palette.Subtle)),
		vspace(4),
		layout.Rigid(field(&a.selfProfileNameEditor, "Name")),
		vspace(10),
		layout.Rigid(a.text("Custom status", 12, a.palette.Subtle)),
		vspace(4),
		layout.Rigid(field(&a.selfProfileStatusEditor, "What's happening?")),
		vspace(14),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.pill(gtx, &a.selfProfileCancelButton, "Cancel", false)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.pill(gtx, &a.selfProfileSaveButton, "Save", true) }),
			)
		}),
	}
	if a.selfProfileError != "" {
		children = append(children, vspace(8), layout.Rigid(a.text(a.selfProfileError, 12, a.palette.Warning)))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}
