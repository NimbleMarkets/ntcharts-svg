package svg

import (
	"strings"
	"testing"
)

func TestCanvasToSVG(t *testing.T) {
	c := NewCanvas(100, 80).
		Background("white").
		SetTitle("test canvas").
		Rect(10, 10, 30, 20, Fill("#f00")).
		Circle(50, 40, 15, Fill("blue").WithOpacity(0.5)).
		Line(0, 0, 100, 80, Stroke("black", 2)).
		Text(50, 70, "hi", Paint{}.WithFont("sans-serif", 10))

	out := c.ToSVG()
	for _, want := range []string{
		`<svg xmlns="http://www.w3.org/2000/svg"`,
		`viewBox="0 0 100 80"`,
		"<title>test canvas</title>",
		`<rect`, `<circle`, `<line`, `<text`,
		`opacity="0.5"`,
		`font-size="10"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("ToSVG() missing %q\n---\n%s", want, out)
		}
	}
}

func TestCanvasEscaping(t *testing.T) {
	c := NewCanvas(50, 50).Text(0, 10, `a<b>&"c`, Paint{})
	out := c.ToSVG()
	if strings.Contains(out, "<b>") {
		t.Errorf("text content not escaped:\n%s", out)
	}
	if !strings.Contains(out, "&lt;b&gt;") {
		t.Errorf("expected escaped text in:\n%s", out)
	}
}

func TestCanvasClampsDimensions(t *testing.T) {
	c := NewCanvas(-5, 0)
	w, h := c.Size()
	if w < 1 || h < 1 {
		t.Errorf("NewCanvas(-5,0) Size = %g×%g, want >= 1×1", w, h)
	}
}

func TestCanvasRasterizes(t *testing.T) {
	c := NewCanvas(120, 60).
		Background("white").
		Circle(60, 30, 25, Fill("#0a0"))
	img, err := c.ToImage(200, 200)
	if err != nil {
		t.Fatalf("ToImage: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != b.Dy()*2 {
		t.Errorf("rasterized canvas bounds %v lost the 2:1 aspect", b)
	}
}

func TestChartBarCanvas(t *testing.T) {
	c := NewChart(320, 200).
		SetTitle("bars").
		SetLabels("a", "b", "c").
		AddSeries("s1", "", 3, 7, 5).
		AddSeries("s2", "#f80", 6, 2, 8).
		BarCanvas()

	out := c.ToSVG()
	if !strings.Contains(out, "<title>bars</title>") {
		t.Errorf("bar chart missing title:\n%s", out)
	}
	if strings.Count(out, "<rect") < 6 {
		t.Errorf("bar chart should have >= 6 bar rects, got:\n%s", out)
	}
	if _, err := RasterizeSVG(c.Bytes(), 400, 400); err != nil {
		t.Errorf("generated bar chart did not rasterize: %v", err)
	}
}

func TestChartLineCanvas(t *testing.T) {
	c := NewChart(320, 200).
		SetLabels("jan", "feb", "mar", "apr").
		AddSeries("temp", "", 1, 4, 9, 16).
		LineCanvas()

	out := c.ToSVG()
	if !strings.Contains(out, "<polyline") {
		t.Errorf("line chart missing polyline:\n%s", out)
	}
	if _, err := RasterizeSVG(c.Bytes(), 400, 400); err != nil {
		t.Errorf("generated line chart did not rasterize: %v", err)
	}
}

func TestEmptyChartIsValid(t *testing.T) {
	// A chart with no series must still emit well-formed, rasterizable
	// SVG rather than panicking on the empty-bounds path.
	c := NewChart(200, 120).BarCanvas()
	if _, err := RasterizeSVG(c.Bytes(), 200, 200); err != nil {
		t.Errorf("empty bar chart did not rasterize: %v", err)
	}
	c = NewChart(200, 120).LineCanvas()
	if _, err := RasterizeSVG(c.Bytes(), 200, 200); err != nil {
		t.Errorf("empty line chart did not rasterize: %v", err)
	}
}

func TestSanitizeForTerminal(t *testing.T) {
	cases := map[string]string{
		"plain":            "plain",
		"tab\there":        "tab here",
		"line\nbreak":      "line break",
		"bell\x07char":     "bellchar",
		"‮rvld":            "rvld", // bidi override stripped
		"emoji \U0001F600": "emoji \U0001F600",
	}
	for in, want := range cases {
		if got := SanitizeForTerminal(in); got != want {
			t.Errorf("SanitizeForTerminal(%q) = %q, want %q", in, got, want)
		}
	}
}
