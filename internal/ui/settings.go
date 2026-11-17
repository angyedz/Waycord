package ui

import "gioui.org/layout"

var settingsNames = [...]string{"Appearance", "Performance", "Audio", "Video", "Updates"}

func (a *App) settings(gtx layout.Context) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(fixedWidth(160, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: 12, Top: 16, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, 0, len(settingsNames)*2)
				for i, title := range settingsNames {
					i, title := i, title
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.clickLabel(gtx, &a.settingsSections[i], title, a.settingsSection == i)
					}), vspace(3))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			})
		})),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if a.settingsSection == 4 {
				return a.updates(gtx)
			}
			return layout.Inset{Top: 16, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				title := settingsNames[a.settingsSection]
				description := ""
				switch a.settingsSection {
				case 0:
					description = "Adjust the spacing in your conversations."
				case 1:
					description = "Performance controls are not available yet."
				case 2:
					description = "Audio devices are not available yet."
				case 3:
					description = "Video devices are not available yet."
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(a.text(title, 24, a.palette.Text)), vspace(8),
					layout.Rigid(a.text(description, 14, a.palette.Muted)), vspace(16),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if a.settingsSection != 0 {
							return layout.Dimensions{}
						}
						label := "Compact messages · Off"
						if a.compactMessages {
							label = "Compact messages · On"
						}
						return a.pill(gtx, &a.compactButton, label, a.compactMessages)
					}),
				)
			})
		}),
	)
}
