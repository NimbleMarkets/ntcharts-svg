package svg

import (
	"os"
	"path/filepath"
	"testing"
)

// repoTestdata locates an SVG fixture regardless of where the test runs.
// Tests usually run with WD = the package dir; the relative search
// covers that case and a run from the repo root.
func repoTestdata(t *testing.T, name string) string {
	t.Helper()
	for _, rel := range []string{
		"testdata/" + name,
		"svg/testdata/" + name,
	} {
		if _, err := os.Stat(rel); err == nil {
			if abs, err := filepath.Abs(rel); err == nil {
				return abs
			}
			return rel
		}
	}
	t.Skipf("testdata fixture %q not found", name)
	return ""
}

func TestLoadSampleFile(t *testing.T) {
	path := repoTestdata(t, "sample.svg")
	m := New(100, 40)

	cmd := m.SetSVG(path)
	if cmd == nil {
		t.Fatal("SetSVG returned nil Cmd")
	}
	m, _ = m.Update(cmd())
	if m.Err() != nil {
		t.Fatalf("load error: %v", m.Err())
	}

	w, h := m.ViewBox()
	if w != 320 || h != 240 {
		t.Errorf("ViewBox() = %g×%g, want 320×240", w, h)
	}
	if m.Title() != "ntcharts-svg sample" {
		t.Errorf("Title() = %q", m.Title())
	}

	m = rasterize(t, m)
	if m.Err() != nil {
		t.Fatalf("render error: %v", m.Err())
	}
	if m.sourceImage == nil {
		t.Fatal("sourceImage nil after rasterizing the sample file")
	}
}

func TestLoadMissingFile(t *testing.T) {
	m := New(40, 10)
	cmd := m.SetSVG("testdata/does-not-exist.svg")
	m, _ = m.Update(cmd())
	if m.Err() == nil {
		t.Fatal("expected error loading a missing file")
	}
}

func TestFileSizeLimit(t *testing.T) {
	path := repoTestdata(t, "sample.svg")
	cfg := Config{Cols: 40, Rows: 10, Limits: Limits{MaxFileBytes: 16}}
	m := NewWithConfig(cfg)
	cmd := m.SetSVG(path)
	m, _ = m.Update(cmd())
	if m.Err() == nil {
		t.Fatal("expected MaxFileBytes to reject the sample file")
	}
}

func TestRoundTripCanvasThroughModel(t *testing.T) {
	canvas := NewChart(300, 180).
		SetTitle("round trip").
		SetLabels("x", "y").
		AddSeries("v", "", 5, 9).
		BarCanvas()

	m := New(80, 24)
	cmd := m.ShowCanvas(canvas)
	if cmd == nil {
		t.Fatal("ShowCanvas returned nil Cmd")
	}
	m, _ = m.Update(cmd())
	if m.Err() != nil {
		t.Fatalf("ShowCanvas load error: %v", m.Err())
	}
	if m.Title() != "round trip" {
		t.Errorf("Title() = %q, want %q", m.Title(), "round trip")
	}
	m = rasterize(t, m)
	if m.sourceImage == nil {
		t.Fatal("generated canvas did not rasterize through the Model")
	}
}
