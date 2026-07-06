package kumi

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

// ══════════════════════════════════════════════════
// .polypack file-type icon
//
// A distinct icon so pack files are visually separate from the PolyForge
// app itself. It's drawn programmatically (a stacked-cards "bundle" mark in
// the brand purple) and generated at first run, so there's no binary asset
// to commit and it always matches the current build.
// ══════════════════════════════════════════════════

// packIconICO returns a Windows .ico (single 256×256 PNG-in-ICO entry, which
// modern Windows renders for file types).
func packIconICO() ([]byte, error) {
	pngBytes, err := packIconPNG()
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	// ICONDIR
	_ = binary.Write(&b, binary.LittleEndian, uint16(0)) // reserved
	_ = binary.Write(&b, binary.LittleEndian, uint16(1)) // type: icon
	_ = binary.Write(&b, binary.LittleEndian, uint16(1)) // count
	// ICONDIRENTRY
	b.WriteByte(0) // width  0 = 256
	b.WriteByte(0) // height 0 = 256
	b.WriteByte(0) // palette
	b.WriteByte(0) // reserved
	_ = binary.Write(&b, binary.LittleEndian, uint16(1))              // color planes
	_ = binary.Write(&b, binary.LittleEndian, uint16(32))             // bpp
	_ = binary.Write(&b, binary.LittleEndian, uint32(len(pngBytes)))  // size
	_ = binary.Write(&b, binary.LittleEndian, uint32(6+16))           // offset
	b.Write(pngBytes)
	return b.Bytes(), nil
}

// packIconPNG renders the icon at 4× and box-downsamples to 256 for AA.
func packIconPNG() ([]byte, error) {
	const n = 256
	const f = 4
	const s = n * f
	hi := image.NewNRGBA(image.Rect(0, 0, s, s))

	// Three offset rounded cards, back (darkest) to front (lightest) — reads
	// as a bundle/collection, clearly not a single app logo. Coordinates are
	// in the supersampled s×s (1024) space.
	type card struct {
		x, y int
		fill color.NRGBA
	}
	const cw, ch, cr = 520, 520, 96
	cards := []card{
		{x: 180, y: 140, fill: color.NRGBA{0x5a, 0x00, 0x96, 0xff}},
		{x: 260, y: 240, fill: color.NRGBA{0x7a, 0x14, 0xc8, 0xff}},
		{x: 340, y: 340, fill: color.NRGBA{0xa3, 0x47, 0xff, 0xff}},
	}
	for _, c := range cards {
		fillRoundRect(hi, c.x, c.y, c.x+cw, c.y+ch, cr, c.fill)
	}
	// Top-edge highlight on the front card for a little depth.
	front := cards[2]
	fillRoundRect(hi, front.x+70, front.y+60, front.x+cw-70, front.y+132, 34, color.NRGBA{0xff, 0xff, 0xff, 0x38})

	// Box-downsample 4×4 → 256×256.
	out := image.NewNRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			var r, g, b, a int
			for dy := 0; dy < f; dy++ {
				for dx := 0; dx < f; dx++ {
					p := hi.NRGBAAt(x*f+dx, y*f+dy)
					r += int(p.R)
					g += int(p.G)
					b += int(p.B)
					a += int(p.A)
				}
			}
			d := f * f
			out.SetNRGBA(x, y, color.NRGBA{uint8(r / d), uint8(g / d), uint8(b / d), uint8(a / d)})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fillRoundRect fills a rounded rectangle (hard edges; AA comes from the
// supersampled downscale) with src over the destination.
func fillRoundRect(img *image.NRGBA, x0, y0, x1, y1, r int, src color.NRGBA) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if !insideRoundRect(x, y, x0, y0, x1, y1, r) {
				continue
			}
			img.SetNRGBA(x, y, overNRGBA(img.NRGBAAt(x, y), src))
		}
	}
}

func insideRoundRect(x, y, x0, y0, x1, y1, r int) bool {
	// Corner circle centers.
	cx, cy := x, y
	switch {
	case x < x0+r && y < y0+r:
		cx, cy = x0+r, y0+r
	case x >= x1-r && y < y0+r:
		cx, cy = x1-r-1, y0+r
	case x < x0+r && y >= y1-r:
		cx, cy = x0+r, y1-r-1
	case x >= x1-r && y >= y1-r:
		cx, cy = x1-r-1, y1-r-1
	default:
		return true // in the straight edges / center
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= r*r
}

// overNRGBA composites src over dst (both non-premultiplied).
func overNRGBA(dst, src color.NRGBA) color.NRGBA {
	sa := float64(src.A) / 255
	da := float64(dst.A) / 255
	outA := sa + da*(1-sa)
	if outA == 0 {
		return color.NRGBA{}
	}
	mix := func(sc, dc uint8) uint8 {
		s := float64(sc) / 255
		d := float64(dc) / 255
		return uint8((s*sa + d*da*(1-sa)) / outA * 255)
	}
	return color.NRGBA{mix(src.R, dst.R), mix(src.G, dst.G), mix(src.B, dst.B), uint8(outA * 255)}
}
