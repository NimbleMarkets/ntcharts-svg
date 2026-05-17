# CHANGELOG

## v0.1.1 (2026-05-17)

 * Add [GitHub pages demo](https://nimblemarkets.github.io/ntcharts-svg/) using [`booba`](https://github.com/NimbleMarkets/go-booba)

## v0.1.0 (2026-05-17)

  * Initial release
  * `svg.Model` Bubble Tea widget with RasterMode and InfoMode
  * Pure-Go SVG rasterization via `oksvg`/`rasterx` (works on native, WASM, WASI)
  * Zoom / pan viewport for inspecting rasterized pixels
  * Immediate-mode `Canvas` SVG generator with bar / line `Chart` helpers
  * `Config.Limits` resource caps for untrusted documents
