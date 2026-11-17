package ui

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func TestServerFeaturesOnlyShowWhenDataExists(t *testing.T) {
	features := availableServerFeatures(mock.Servers[0])
	if !containsFeature(features, serverFeatureGuide) || !containsFeature(features, serverFeatureEvents) || !containsFeature(features, serverFeatureRoles) {
		t.Fatalf("first server is missing available features: %v", features)
	}
	if containsFeature(features, serverFeatureBoosts) {
		t.Fatal("server without boost data showed a boosts feature")
	}
	second := availableServerFeatures(mock.Servers[1])
	if !containsFeature(second, serverFeatureBoosts) || containsFeature(second, serverFeatureEvents) || containsFeature(second, serverFeatureGuide) {
		t.Fatalf("second server did not respect its available feature data: %v", second)
	}
}

func TestServerFeatureActionsUpdateOnlyLocalState(t *testing.T) {
	app := New()
	app.toggleEventInterest(0, 0)
	app.toggleGuideStep(0, 1)
	if !app.eventInterested[serverFeatureItemKey(0, 0)] || !app.guideCompleted[serverFeatureItemKey(0, 1)] {
		t.Fatal("event interest and guide completion were not stored locally")
	}
	app.toggleEventInterest(0, 0)
	app.toggleGuideStep(0, 1)
	if app.eventInterested[serverFeatureItemKey(0, 0)] || app.guideCompleted[serverFeatureItemKey(0, 1)] {
		t.Fatal("local event and guide controls did not toggle off")
	}
}

func TestServerFeatureNavigationOpensAndReturnsToChat(t *testing.T) {
	app := New()
	clickButton(t, app, func(gtx layout.Context) layout.Dimensions {
		return app.clickLabel(gtx, &app.serverFeatureButtons[0][serverFeatureEvents], "Events", false)
	})
	if app.activeServerFeature != serverFeatureEvents {
		t.Fatalf("events row did not open its screen: %d", app.activeServerFeature)
	}
	clickButton(t, app, func(gtx layout.Context) layout.Dimensions {
		return app.pill(gtx, &app.chatButton, "Back", false)
	})
	if app.activeServerFeature != serverFeatureNone {
		t.Fatalf("back control did not return to chat: %d", app.activeServerFeature)
	}
}

func TestAvailableServerFeatureScreensRender(t *testing.T) {
	for serverIndex, server := range mock.Servers {
		for _, feature := range availableServerFeatures(server) {
			app := New()
			app.activeServer, app.activeServerFeature = serverIndex, feature
			var ops op.Ops
			gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(720, 620)}}
			dimensions := app.serverFeatureScreen(gtx)
			if dimensions.Size.X == 0 || dimensions.Size.Y == 0 || dimensions.Size.X > 720 || dimensions.Size.Y > 620 {
				t.Fatalf("server %d feature %d rendered outside its bounds: %+v", serverIndex, feature, dimensions)
			}
		}
	}
}

func clickButton(t *testing.T, app *App, draw layout.Widget) {
	t.Helper()
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(340, 48)}}
	draw(gtx)
	router.Frame(&ops)
	router.Queue(
		pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 18)},
		pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(20, 18)},
	)
	app.process(gtx)
}

func containsFeature(features []int, target int) bool {
	for _, feature := range features {
		if feature == target {
			return true
		}
	}
	return false
}
