// Copyright © 2016 Steve Streeting
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package recolour

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaletteImageDimensions(t *testing.T) {
	w, h := getPaletteImageDimensions(3)
	assert.EqualValues(t, 4, w)
	assert.EqualValues(t, 1, h)

	w, h = getPaletteImageDimensions(12)
	assert.EqualValues(t, 16, w)
	assert.EqualValues(t, 1, h)

	w, h = getPaletteImageDimensions(32)
	assert.EqualValues(t, 32, w)
	assert.EqualValues(t, 1, h)

	w, h = getPaletteImageDimensions(122)
	assert.EqualValues(t, 128, w)
	assert.EqualValues(t, 1, h)

	w, h = getPaletteImageDimensions(200)
	assert.EqualValues(t, 256, w)
	assert.EqualValues(t, 1, h)

	w, h = getPaletteImageDimensions(256)
	assert.EqualValues(t, 256, w)
	assert.EqualValues(t, 1, h)

	w, h = getPaletteImageDimensions(257)
	assert.EqualValues(t, 256, w)
	assert.EqualValues(t, 2, h)

	w, h = getPaletteImageDimensions(530)
	assert.EqualValues(t, 256, w)
	assert.EqualValues(t, 4, h)

	w, h = getPaletteImageDimensions(1700)
	assert.EqualValues(t, 256, w)
	assert.EqualValues(t, 8, h)
}
