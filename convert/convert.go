// this package takes an image as input, and coverts it to an ASCII string
package convert

import (
	"image"
	"strings"
	"sync"
)

// CharStep
// colours we're getting fro the image are in the 0 - 0xffff range per channel (RGB)
// we need to figure out how to map that onto a 29 rune slice, values 0-N will be represented by
// a single character, the next N colours move ahead in the slice and so on, so each character represents
// this many colours: (0xFFFF * 3) / 29 or (max value per channel * number of channels) / number of chars
// this is a float to be as precise as possible
const CharStep = float64((65535.0 * 3.0) / 29.0)

// ASCIIChars characters we'll use to build up or image
var ASCIIChars = []rune("Ñ@#W$9876543210?!abc;:+=-,._ ")

// PixelChar the character for a given pixel in the image
type PixelChar struct {
	x, y int
	char rune
}

// ImgToASCII converts an image to a string. By default, ligher colours will be represented by smaller characters
// all the way down to white being shown as a space. pass in true for negative will swap this around, where spaces represent black pixels
// and vice-versa
// invert will mirror the image (useful for webcam)
func ImgToASCII(img image.Image, negative, invert bool) string {
	max := img.Bounds().Max
	wg := sync.WaitGroup{}
	wg.Add(max.Y)
	done := make(chan struct{})             // the routine that will populate the slice  will let us know when it's done with this
	ch := make(chan PixelChar, max.Y+max.X) // buffer enough for first pixels of each row + 1 column
	matrix := make([][]rune, max.Y)         // matrix[height][width]
	// start waiting for data
	go func() {
		for pc := range ch {
			i := pc.x
			if invert {
				i = len(matrix[pc.y]) - i - 1
			}
			matrix[pc.y][i] = pc.char
		}
		close(done)
	}()
	for y := 0; y < max.Y; y++ {
		matrix[y] = make([]rune, max.X) // initialise each column
		go convertRow(&wg, ch, img, y, negative)
	}
	wg.Wait()
	close(ch)
	<-done
	// OK, our slice is populated, convert to string:

	chunks := make([]string, 0, len(matrix))
	for _, r := range matrix {
		// this package is not supposed to trim trailing spaces. It faithfully converts all pixels, and returns them
		// the caller may decide to trim
		chunks = append(chunks, string(r))
		// chunks = append(chunks, strings.TrimRight(string(r), " "))
	}
	// return entire image as a string
	return strings.Join(chunks, "\n")
}

func convertRow(wg *sync.WaitGroup, ch chan<- PixelChar, img image.Image, y int, reverse bool) {
	max := img.Bounds().Max.X
	cLen := len(ASCIIChars)
	for x := 0; x < max; x++ {
		// alpha is already applied, so we can just ignore it
		i := 0
		r, g, b, a := img.At(x, y).RGBA()
		if a == 0 {
			// alpha on max, space character
			i = cLen - 1
		} else {
			i = int(float64(r+g+b) / CharStep)
			if !reverse {
				i = cLen - i
			}
		}
		ch <- PixelChar{
			char: ASCIIChars[i%cLen],
			x:    x,
			y:    y,
		}
	}
	wg.Done()
}
