package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/ntcharts-svg/svg"
	"github.com/NimbleMarkets/ntcharts/v2/picture"
)

func TestDemoPrefersKittyWhenAvailable(t *testing.T) {
	previous := picture.KittySupported()
	t.Cleanup(func() { picture.ForceKittyCapability(previous) })
	picture.ForceKittyCapability(picture.KittyCapabilityUnknown)
	m := initialModel(svg.Config{})
	update := func(msg tea.Msg) {
		next, _ := m.Update(msg)
		m = next.(model)
	}
	for _, capability := range []picture.KittyCapability{
		picture.KittyCapabilityUnknown, picture.KittyCapabilityUnsupported,
	} {
		picture.ForceKittyCapability(capability)
		update(nil)
		if m.sv.RenderMode() != svg.RenderGlyph {
			t.Fatal("selected Kitty without confirmed support")
		}
	}
	// A positive reply arriving after the probe timeout must still select Kitty.
	picture.ForceKittyCapability(picture.KittyCapabilitySupported)
	update(nil)
	if m.sv.RenderMode() != svg.RenderKitty {
		t.Fatal("did not select Kitty after support was confirmed")
	}
	update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	update(nil)
	if m.sv.RenderMode() != svg.RenderGlyph {
		t.Fatal("automatic selection overrode the user's glyph choice")
	}

	// Manual input while detection is pending also takes precedence.
	picture.ForceKittyCapability(picture.KittyCapabilityUnknown)
	m = initialModel(svg.Config{})
	update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	picture.ForceKittyCapability(picture.KittyCapabilitySupported)
	update(nil)
	if m.sv.RenderMode() != svg.RenderGlyph {
		t.Fatal("late detection overrode a manual choice")
	}
}
