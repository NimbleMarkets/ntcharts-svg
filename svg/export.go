// svg/export.go: turning SVG markup — loaded files or generated
// Canvases — into raster images and files, plus the bridge that feeds a
// Canvas into the Model.

package svg

import (
	"fmt"
	"image"
	"image/png"
	"os"

	tea "charm.land/bubbletea/v2"
)

// RasterizeSVG rasterizes raw SVG bytes into an RGBA image fitted to
// maxW × maxH pixels while preserving aspect ratio. It uses the same
// pure-Go oksvg/rasterx backend as the widget's default renderer, so it
// works on every build target.
func RasterizeSVG(data []byte, maxW, maxH int) (image.Image, error) {
	r, err := DefaultRendererFactory()("export", data)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return r.Render(maxW, maxH)
}

// WriteSVG writes the canvas to path as a standalone .svg file (0644).
func (c *Canvas) WriteSVG(path string) error {
	return os.WriteFile(path, c.Bytes(), 0o644)
}

// ToImage rasterizes the canvas into an RGBA image fitted to
// maxW × maxH pixels.
func (c *Canvas) ToImage(maxW, maxH int) (image.Image, error) {
	return RasterizeSVG(c.Bytes(), maxW, maxH)
}

// WritePNG rasterizes the canvas and writes it to path as a PNG (0644).
// maxW / maxH cap the output resolution.
func (c *Canvas) WritePNG(path string, maxW, maxH int) error {
	img, err := c.ToImage(maxW, maxH)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png %q: %w", path, err)
	}
	return nil
}

// ShowCanvas loads a generated Canvas into the viewer, exactly as if its
// SVG markup had been opened from disk. The returned Cmd performs the
// metadata scan and rasterize off the UI goroutine. Pair with NewCanvas
// / NewChart to display generated charts inside a Bubble Tea program.
func (m *Model) ShowCanvas(c *Canvas) tea.Cmd {
	name := "canvas.svg"
	if c.title != "" {
		name = c.title
	}
	return m.SetSVGData(name, c.Bytes())
}
