// svg/canvas.go: an immediate-mode vector drawing surface that emits
// clean, standalone SVG.
//
// Canvas is a self-contained pure-Go SVG writer — no external rendering
// dependency, so it compiles for every target the widget supports. Draw
// with the chainable primitive methods (Line, Rect, Circle, Path,
// Text…), then ToSVG / Bytes for the markup or ToImage / WriteSVG (see
// export.go) to rasterize or persist. ShowCanvas on a Model wires a
// finished Canvas straight into the viewer.
//
// Higher-level chart helpers (bar / line / axes) build on these
// primitives in chart.go.

package svg

import (
	"fmt"
	"strconv"
	"strings"
)

// Point is an (x, y) coordinate in canvas user units.
type Point struct{ X, Y float64 }

// Paint describes how a shape is filled and stroked. The zero value is
// a usable default: solid black fill, no stroke, fully opaque. Build
// variants fluently with the With* methods or the Fill / Stroke
// constructors.
type Paint struct {
	Fill        string  // CSS color, or "none"; "" means solid black
	Stroke      string  // CSS color; "" or "none" means no stroke
	StrokeWidth float64 // stroke width in user units; <=0 becomes 1 when stroked
	Opacity     float64 // 0 is treated as 1 (opaque); set 0<o<1 for translucency
	Dash        string  // stroke-dasharray value, e.g. "4 2"
	LineCap     string  // "butt" | "round" | "square"
	FontFamily  string  // text family; "" means the SVG default
	FontSize    float64 // text size in user units
	TextAnchor  string  // "start" | "middle" | "end"
}

// Fill returns a Paint that fills with color and draws no stroke.
func Fill(color string) Paint { return Paint{Fill: color} }

// Stroke returns a Paint that strokes with color at the given width and
// does not fill ("none").
func Stroke(color string, width float64) Paint {
	return Paint{Fill: "none", Stroke: color, StrokeWidth: width}
}

// WithFill returns a copy of p with the fill color replaced.
func (p Paint) WithFill(color string) Paint { p.Fill = color; return p }

// WithStroke returns a copy of p with the stroke color and width set.
func (p Paint) WithStroke(color string, width float64) Paint {
	p.Stroke, p.StrokeWidth = color, width
	return p
}

// WithOpacity returns a copy of p with opacity set (0 < o <= 1).
func (p Paint) WithOpacity(o float64) Paint { p.Opacity = o; return p }

// WithFont returns a copy of p with the text family and size set.
func (p Paint) WithFont(family string, size float64) Paint {
	p.FontFamily, p.FontSize = family, size
	return p
}

// WithAnchor returns a copy of p with the text anchor set.
func (p Paint) WithAnchor(anchor string) Paint { p.TextAnchor = anchor; return p }

// Canvas is an immediate-mode SVG drawing surface. Construct with
// NewCanvas; every draw method appends an element and returns the
// Canvas for chaining. Canvas is not safe for concurrent use.
type Canvas struct {
	w, h  float64
	bg    string
	title string
	elems []string
}

// NewCanvas returns a Canvas measuring width × height user units. Non-
// positive dimensions are clamped to 1 so the emitted SVG is always
// well-formed.
func NewCanvas(width, height float64) *Canvas {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	return &Canvas{w: width, h: height}
}

// Size reports the canvas dimensions in user units.
func (c *Canvas) Size() (width, height float64) { return c.w, c.h }

// Background sets a solid background fill drawn behind every element.
// Pass "" to clear it (the default — a transparent canvas).
func (c *Canvas) Background(color string) *Canvas { c.bg = color; return c }

// SetTitle sets the document <title>, surfaced by the viewer's InfoMode.
func (c *Canvas) SetTitle(title string) *Canvas { c.title = title; return c }

// Line draws a straight segment from (x1,y1) to (x2,y2). A Paint with no
// stroke defaults to a 1-unit black stroke so the line is always
// visible.
func (c *Canvas) Line(x1, y1, x2, y2 float64, p Paint) *Canvas {
	if p.Stroke == "" || p.Stroke == "none" {
		p.Stroke, p.StrokeWidth = "black", maxF(p.StrokeWidth, 1)
	}
	c.elems = append(c.elems, fmt.Sprintf(`<line x1="%s" y1="%s" x2="%s" y2="%s"%s/>`,
		num(x1), num(y1), num(x2), num(y2), p.attrs("none")))
	return c
}

// Rect draws an axis-aligned rectangle with its top-left corner at (x,y).
func (c *Canvas) Rect(x, y, width, height float64, p Paint) *Canvas {
	c.elems = append(c.elems, fmt.Sprintf(`<rect x="%s" y="%s" width="%s" height="%s"%s/>`,
		num(x), num(y), num(width), num(height), p.attrs("black")))
	return c
}

// RoundRect draws a rectangle with corner radii (rx, ry).
func (c *Canvas) RoundRect(x, y, width, height, rx, ry float64, p Paint) *Canvas {
	c.elems = append(c.elems, fmt.Sprintf(
		`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" ry="%s"%s/>`,
		num(x), num(y), num(width), num(height), num(rx), num(ry), p.attrs("black")))
	return c
}

// Circle draws a circle of radius r centered at (cx,cy).
func (c *Canvas) Circle(cx, cy, r float64, p Paint) *Canvas {
	c.elems = append(c.elems, fmt.Sprintf(`<circle cx="%s" cy="%s" r="%s"%s/>`,
		num(cx), num(cy), num(r), p.attrs("black")))
	return c
}

// Ellipse draws an ellipse with radii (rx,ry) centered at (cx,cy).
func (c *Canvas) Ellipse(cx, cy, rx, ry float64, p Paint) *Canvas {
	c.elems = append(c.elems, fmt.Sprintf(`<ellipse cx="%s" cy="%s" rx="%s" ry="%s"%s/>`,
		num(cx), num(cy), num(rx), num(ry), p.attrs("black")))
	return c
}

// Polyline draws an open path through the given points. With no stroke
// the Paint defaults to a 1-unit black stroke and no fill.
func (c *Canvas) Polyline(pts []Point, p Paint) *Canvas {
	if p.Stroke == "" || p.Stroke == "none" {
		p.Stroke, p.StrokeWidth = "black", maxF(p.StrokeWidth, 1)
	}
	if p.Fill == "" {
		p.Fill = "none"
	}
	c.elems = append(c.elems, fmt.Sprintf(`<polyline points="%s"%s/>`,
		pointList(pts), p.attrs("none")))
	return c
}

// Polygon draws a closed shape through the given points.
func (c *Canvas) Polygon(pts []Point, p Paint) *Canvas {
	c.elems = append(c.elems, fmt.Sprintf(`<polygon points="%s"%s/>`,
		pointList(pts), p.attrs("black")))
	return c
}

// Path draws a raw SVG path-data string ("M0 0 L10 10 …"). The caller
// owns the geometry; the data string is attribute-escaped but not
// otherwise validated.
func (c *Canvas) Path(d string, p Paint) *Canvas {
	c.elems = append(c.elems, fmt.Sprintf(`<path d="%s"%s/>`, escAttr(d), p.attrs("black")))
	return c
}

// Text draws a text run anchored at (x,y). The string content is XML-
// escaped. Font family / size / anchor come from the Paint.
func (c *Canvas) Text(x, y float64, s string, p Paint) *Canvas {
	attrs := p.attrs("black")
	if p.FontFamily != "" {
		attrs += ` font-family="` + escAttr(p.FontFamily) + `"`
	}
	if p.FontSize > 0 {
		attrs += ` font-size="` + num(p.FontSize) + `"`
	}
	if p.TextAnchor != "" {
		attrs += ` text-anchor="` + escAttr(p.TextAnchor) + `"`
	}
	c.elems = append(c.elems, fmt.Sprintf(`<text x="%s" y="%s"%s>%s</text>`,
		num(x), num(y), attrs, escText(s)))
	return c
}

// Raw appends a pre-formed SVG fragment verbatim. An escape hatch for
// elements the typed API doesn't cover; the caller is responsible for
// the fragment's well-formedness.
func (c *Canvas) Raw(fragment string) *Canvas {
	c.elems = append(c.elems, fragment)
	return c
}

// ToSVG renders the canvas to a standalone SVG document string.
func (c *Canvas) ToSVG() string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s">`,
		num(c.w), num(c.h), num(c.w), num(c.h)))
	b.WriteByte('\n')
	if c.title != "" {
		b.WriteString("  <title>" + escText(c.title) + "</title>\n")
	}
	if c.bg != "" {
		b.WriteString(fmt.Sprintf("  <rect x=\"0\" y=\"0\" width=\"%s\" height=\"%s\" fill=\"%s\"/>\n",
			num(c.w), num(c.h), escAttr(c.bg)))
	}
	for _, e := range c.elems {
		b.WriteString("  ")
		b.WriteString(e)
		b.WriteByte('\n')
	}
	b.WriteString("</svg>\n")
	return b.String()
}

// Bytes renders the canvas to SVG as a byte slice.
func (c *Canvas) Bytes() []byte { return []byte(c.ToSVG()) }

// attrs builds the shared presentation attributes for one element.
// defaultFill is used when Paint.Fill is empty.
func (p Paint) attrs(defaultFill string) string {
	var b strings.Builder
	fill := p.Fill
	if fill == "" {
		fill = defaultFill
	}
	b.WriteString(` fill="` + escAttr(fill) + `"`)
	if p.Stroke != "" && p.Stroke != "none" {
		b.WriteString(` stroke="` + escAttr(p.Stroke) + `"`)
		sw := p.StrokeWidth
		if sw <= 0 {
			sw = 1
		}
		b.WriteString(` stroke-width="` + num(sw) + `"`)
		if p.LineCap != "" {
			b.WriteString(` stroke-linecap="` + escAttr(p.LineCap) + `"`)
		}
		if p.Dash != "" {
			b.WriteString(` stroke-dasharray="` + escAttr(p.Dash) + `"`)
		}
	}
	// Opacity 0 is treated as the opaque default so the zero-value
	// Paint draws something; only an explicit 0<o<1 emits the attr.
	if p.Opacity > 0 && p.Opacity < 1 {
		b.WriteString(` opacity="` + num(p.Opacity) + `"`)
	}
	return b.String()
}

// num formats a float with the shortest exact decimal representation,
// dropping trailing zeros so the markup stays tidy.
func num(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// pointList formats points as the "x,y x,y …" form used by polyline /
// polygon.
func pointList(pts []Point) string {
	var b strings.Builder
	for i, p := range pts {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(num(p.X))
		b.WriteByte(',')
		b.WriteString(num(p.Y))
	}
	return b.String()
}

// escAttr escapes a string for use inside a double-quoted XML attribute.
func escAttr(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

// escText escapes a string for use as XML text content.
func escText(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
