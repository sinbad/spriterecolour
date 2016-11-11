package recolour

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	colorful "github.com/lucasb-eyer/go-colorful"
	// This causes the codecs to be loaded
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

var EPSILON float64 = 0.00000001

func floatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}

type UniqueColour struct {
	RGBA   color.RGBA
	colour colorful.Color
	// Store an index so that references in map know final position in list
	Index int
}

func sortColours(inlist []*UniqueColour) []*UniqueColour {
	// First generate a distance map
	// Some duplication here A->B and B->A both stored but live with for simplicity
	n := len(inlist)
	distances := make([][]float64, n)
	for fromN := 0; fromN < n; fromN++ {
		distances[fromN] = make([]float64, n)
		for toN := 0; toN < n; toN++ {
			if toN == fromN {
				distances[fromN][toN] = 0.0
			} else {
				distances[fromN][toN] = inlist[fromN].colour.DistanceLab(inlist[toN].colour)
			}
		}
	}

	visited := make([]bool, n)

	// Now do a nearest neighbour walk
	outList := make([]*UniqueColour, 0, n)
	// Arbitrarily pick the first colour
	currentNode := 0
	outList = append(outList, inlist[currentNode])
	visited[currentNode] = true
	for i := 1; i < n; i++ {
		minDistance := float64(99999999999999.9)
		bestNode := -1
		for j := 0; j < n; j++ {
			if visited[j] {
				continue
			}
			dist := distances[currentNode][j]
			if dist < minDistance {
				minDistance = dist
				bestNode = j
			}
		}
		if bestNode == -1 {
			fmt.Fprintf(os.Stderr, "Ran out of colours to sort, this is a bug")
			break
		}
		currentNode = bestNode
		newcol := inlist[currentNode]
		newcol.Index = i
		outList = append(outList, newcol)
		visited[currentNode] = true
	}
	return outList
}

// Why the hell doesn't image/color have  path for this? Only the reverse
func colourTo8BitRGBA(c color.Color) color.RGBA {
	// color.Color.RGBA is 0-65535 even from 8-bit channel images because of course it is
	// Also premultiplied alpha but we'll preserve that if present
	r, g, b, a := c.RGBA()
	return color.RGBA{
		uint8((float64(r) / 65535.0) * 255.0),
		uint8((float64(g) / 65535.0) * 255.0),
		uint8((float64(b) / 65535.0) * 255.0),
		uint8((float64(a) / 65535.0) * 255.0),
	}

}

func colourTo8BitPaletteRGBA(c color.Color) color.RGBA {
	cout := colourTo8BitRGBA(c)
	// Palette entries must be solid
	cout.A = 255
	return cout
}

// Generate reads an input sprite texture and generates a reference sprite file,
// and optionally a palette texture. If outPaletteTexture is supplied then the
// R channel will be rescaled based on the palette texture size; otherwise it
// will be the colour index (and limited to 256 colours)
func Generate(imagePath, outImagePath, outPaletteTexture string) ([]color.RGBA, error) {

	f, err := os.OpenFile(imagePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return GenerateFromImage(img, outImagePath, outPaletteTexture)
}

// GenerateFromImage reads an image and generates a reference sprite file,
// and optionally a palette texture. If outPaletteTexture is supplied then the
// R channel will be rescaled based on the palette texture size; otherwise it
// will be the colour index (and limited to 256 colours)
func GenerateFromImage(img image.Image, outImagePath, outPaletteTexture string) ([]color.RGBA, error) {
	bounds := img.Bounds()
	// Record of what colours are present
	colourMap := make(map[color.RGBA]*UniqueColour)
	// Build colour list as we go so ordering based on encountered pixels is deterministic
	// If we use the map to generate later, ordering is random
	colourList := make([]*UniqueColour, 0)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Go colours are alpha-premultiplied and uint32's with 65535 range: weird
			// We want NON alpha premultiplied by default (internally could be premultiplied)
			p := colourTo8BitPaletteRGBA(img.At(x, y))

			if _, ok := colourMap[p]; !ok {
				cfcol := colorful.Color{float64(p.R) / 255.0, float64(p.G) / 255.0, float64(p.B) / 255.0}
				col := &UniqueColour{p, cfcol, len(colourList)}
				colourMap[p] = col
				colourList = append(colourList, col)
			}
		}
	}

	generatePaletteTexture := len(outPaletteTexture) > 0

	if len(colourMap) > 256 && !generatePaletteTexture {
		return nil, fmt.Errorf("Sorry, sprite contains too many colours for shader parameters (>256)")
	}
	if len(colourMap) > 65536 {
		return nil, fmt.Errorf("Sorry, sprite contains too many colours (>65536)")
	}

	// Re-order the colours by proximity so easier to edit
	// Sort, the swap function will swap indexes
	colourList = sortColours(colourList)

	paletteWidth, paletteHeight := getPaletteImageDimensions(len(colourList))
	rescaler := NewTexCoordRescale(paletteWidth, paletteHeight)

	// Now generate the sprite output
	outSprite := image.NewRGBA(image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			inpix := colourTo8BitRGBA(img.At(x, y))
			inpixLookup := colourTo8BitPaletteRGBA(img.At(x, y))

			// Should never fail but just don't write pixel if it does
			if col, ok := colourMap[inpixLookup]; ok {
				var red, green uint8
				// Red channel = colour index U
				red = uint8(col.Index & 0x0000FFFF)
				if generatePaletteTexture {
					// Need to scale UVs from 0-256 to 0-paletteSize
					red = rescaler.rescaleHorizontal(red)
				}
				// Green channel = colour index V
				green = uint8(col.Index >> 16)
				if generatePaletteTexture {
					// Need to scale UVs from 0-256 to 0-paletteSize
					green = rescaler.rescaleVertical(green)
				}
				// Blue channel = unused for now
				outSprite.Set(x, y, color.RGBA{red, green, 0, inpix.A})
			}
		}
	}
	of, err := os.OpenFile(outImagePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	err = png.Encode(of, outSprite)
	of.Close()
	if err != nil {
		return nil, err
	}

	// Now write palette texture & build return
	palette := make([]color.RGBA, 0, len(colourList))
	if len(outPaletteTexture) > 0 {
		outPalette := image.NewRGBA(image.Rect(0, 0, paletteWidth, paletteHeight))
		x := 0
		y := 0
		for n := 0; n < len(colourList); n++ {
			outPalette.SetRGBA(x, y, colourList[n].RGBA)
			palette = append(palette, colourList[n].RGBA)
			x++
			if x == paletteWidth {
				x = 0
				y++
			}
		}

		opf, err := os.OpenFile(outPaletteTexture, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		err = png.Encode(opf, outPalette)
		opf.Close()
		if err != nil {
			return nil, err
		}

	}

	return palette, nil
}

// Texture coordinate rescaler with saved factors
type TexCoordRescale struct {
	rescaleXMul float64
	rescaleXOff float64
	rescaleYMul float64
	rescaleYOff float64
}

func NewTexCoordRescale(paletteWidth, paletteHeight int) TexCoordRescale {
	xMul := 256.0 / float64(paletteWidth)
	xOff := xMul * 0.5 // to ensure we target middle of texel
	yMul := 256.0 / float64(paletteWidth)
	yOff := yMul * 0.5
	return TexCoordRescale{xMul, xOff, yMul, yOff}
}

func (t *TexCoordRescale) rescaleHorizontal(col uint8) uint8 {
	return uint8(float64(col)*t.rescaleXMul + t.rescaleXOff)
}
func (t *TexCoordRescale) rescaleVertical(col uint8) uint8 {
	return uint8(float64(col)*t.rescaleYMul + t.rescaleYOff)
}

func nextPowerOfTwo(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

func getPaletteImageDimensions(numColours int) (width, height int) {
	width = 256
	height = 1
	if numColours > 256 {
		height = nextPowerOfTwo(int(math.Ceil(float64(numColours) / 256.0)))
	} else if numColours <= 128 {
		width = nextPowerOfTwo(numColours)
	}
	return
}
