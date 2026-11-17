package ui

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/widget/material"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

const serverFeatureNone = -1

const (
	serverFeatureGuide = iota
	serverFeatureEvents
	serverFeatureRoles
	serverFeatureBoosts
	serverFeatureCount
)

var serverFeatureNames = [...]string{"Server guide", "Events", "Channels & roles", "Server boosts"}

func availableServerFeatures(server mock.Server) []int {
	features := make([]int, 0, serverFeatureCount)
	for feature := range serverFeatureNames {
		if serverHasFeature(server, feature) {
			features = append(features, feature)
		}
	}
	return features
}

func serverHasFeature(server mock.Server, feature int) bool {
	switch feature {
	case serverFeatureGuide:
		return server.Guide != nil && len(server.Guide.Steps) > 0
	case serverFeatureEvents:
		return len(server.Events) > 0
	case serverFeatureRoles:
		return len(server.Roles) > 0
	case serverFeatureBoosts:
		return server.Boosts != nil && server.Boosts.NextLevel > 0
	default:
		return false
	}
}

func serverFeatureItemKey(server, item int) string {
	return fmt.Sprintf("%d:%d", server, item)
}

func (a *App) toggleGuideStep(server, step int) {
	if server < 0 || server >= len(mock.Servers) || mock.Servers[server].Guide == nil || step < 0 || step >= len(mock.Servers[server].Guide.Steps) {
		return
	}
	key := serverFeatureItemKey(server, step)
	a.guideCompleted[key] = !a.guideCompleted[key]
}

func (a *App) toggleEventInterest(server, event int) {
	if server < 0 || server >= len(mock.Servers) || event < 0 || event >= len(mock.Servers[server].Events) {
		return
	}
	key := serverFeatureItemKey(server, event)
	a.eventInterested[key] = !a.eventInterested[key]
}

func (a *App) serverFeatureRows() []layout.FlexChild {
	var children []layout.FlexChild
	for _, feature := range availableServerFeatures(mock.Servers[a.activeServer]) {
		feature := feature
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.clickLabel(gtx, &a.serverFeatureButtons[a.activeServer][feature], serverFeatureNames[feature], a.activeServerFeature == feature)
			}),
			vspace(3),
		)
	}
	return children
}

func (a *App) serverFeatureScreen(gtx layout.Context) layout.Dimensions {
	server := mock.Servers[a.activeServer]
	if a.activeServerFeature < 0 || a.activeServerFeature >= serverFeatureCount || !serverHasFeature(server, a.activeServerFeature) {
		return layout.Dimensions{}
	}
	return layout.Inset{Top: 22, Left: 28, Right: 28, Bottom: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(a.text(serverFeatureNames[a.activeServerFeature], 23, a.palette.Text)),
			vspace(5),
			layout.Rigid(a.text(server.Name, 13, a.palette.Subtle)),
			vspace(18),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				switch a.activeServerFeature {
				case serverFeatureGuide:
					return a.serverGuideContent(gtx, server)
				case serverFeatureEvents:
					return a.serverEventsContent(gtx, server)
				case serverFeatureRoles:
					return a.serverRolesContent(gtx, server)
				case serverFeatureBoosts:
					return a.serverBoostsContent(gtx, server)
				default:
					return layout.Dimensions{}
				}
			}),
		)
	})
}

func (a *App) serverGuideContent(gtx layout.Context, server mock.Server) layout.Dimensions {
	children := []layout.FlexChild{layout.Rigid(a.text("A few quick steps to get comfortable here.", 14, a.palette.Muted)), vspace(16)}
	for step, title := range server.Guide.Steps {
		step, title := step, title
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			key := serverFeatureItemKey(a.activeServer, step)
			checked := a.guideCompleted[key]
			marker := "○"
			if checked {
				marker = "✓"
			}
			return a.clickLabel(gtx, &a.guideStepButtons[a.activeServer][step], marker+"   "+title, checked)
		}), vspace(5))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (a *App) serverEventsContent(gtx layout.Context, server mock.Server) layout.Dimensions {
	children := []layout.FlexChild{layout.Rigid(a.text("Upcoming in this community", 14, a.palette.Muted)), vspace(14)}
	for index, event := range server.Events {
		index, event := index, event
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			key := serverFeatureItemKey(a.activeServer, index)
			going := a.eventInterested[key]
			label := "Interested"
			if going {
				label = "Going ✓"
			}
			return rounded(gtx, a.palette.Surface, 14, padding(layout.Inset{Left: 16, Right: 12, Top: 12, Bottom: 12}, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(a.text(event.Name, 16, a.palette.Text)),
							layout.Rigid(a.text(event.When+" · "+event.Location, 12, a.palette.Muted)),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.pill(gtx, &a.eventInterestButtons[a.activeServer][index], label, going)
					}),
				)
			}))
		}), vspace(8))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (a *App) serverRolesContent(gtx layout.Context, server mock.Server) layout.Dimensions {
	children := []layout.FlexChild{layout.Rigid(a.text("Roles in this server", 14, a.palette.Muted)), vspace(14)}
	for _, role := range server.Roles {
		role := role
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return rounded(gtx, a.palette.Surface, 12, padding(layout.Inset{Left: 14, Right: 14, Top: 11, Bottom: 11}, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(fixedWidth(10, fixedHeight(10, func(gtx layout.Context) layout.Dimensions {
						return rounded(gtx, a.palette.AccentDim, 5, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
					}))),
					hspace(10),
					layout.Flexed(1, a.text(role.Name, 14, a.palette.Text)),
					layout.Rigid(a.text(fmt.Sprintf("%d members", role.Members), 12, a.palette.Subtle)),
				)
			}))
		}), vspace(6))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (a *App) serverBoostsContent(gtx layout.Context, server mock.Server) layout.Dimensions {
	boosts := server.Boosts
	if boosts == nil || boosts.NextLevel <= 0 {
		return layout.Dimensions{}
	}
	progress := float32(boosts.Count) / float32(boosts.NextLevel)
	return rounded(gtx, a.palette.Surface, 16, padding(layout.Inset{Left: 18, Right: 18, Top: 16, Bottom: 16}, func(gtx layout.Context) layout.Dimensions {
		bar := material.ProgressBar(a.theme, progress)
		bar.Color, bar.TrackColor, bar.Height = a.palette.Accent, a.palette.Surface2, 8
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(a.text(fmt.Sprintf("Level %d", boosts.Level), 21, a.palette.Text)),
			vspace(5),
			layout.Rigid(a.text(fmt.Sprintf("%d of %d boosts toward the next level", boosts.Count, boosts.NextLevel), 13, a.palette.Muted)),
			vspace(12),
			layout.Rigid(bar.Layout),
			vspace(12),
			layout.Rigid(a.text(fmt.Sprintf("Supported by %d members", boosts.Supporters), 12, a.palette.Subtle)),
		)
	}))
}
