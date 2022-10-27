// the job of this package is to simply rescale an image to a given resolution
package scale

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// Mode the scaling algorithm to use
type Mode uint32

// ScaleOpts are the options that can be specified to scale an image
type ScaleOpts struct {
	Width, Height uint
	Factor        float64
	Mode          Mode
}

const (
	NearestNeighbourScaling Mode = iota
	ApproxBilinearScaling
	BilinearScaling
	CatmullRomScaling
)

// for human-readible representation
var (
	scaleStr = map[Mode]string{
		NearestNeighbourScaling: "Nearest Neighbour",
		ApproxBilinearScaling:   "Approximate Bilinear",
		BilinearScaling:         "Bilinear",
		CatmullRomScaling:       "CatmullRom",
	}

	// OrderLHQ the order of scaling algorithms from low quality, high speed to high quality, low speed
	OrderLHQ = []Mode{
		NearestNeighbourScaling,
		ApproxBilinearScaling,
		BilinearScaling,
		CatmullRomScaling,
	}

	supportedTypes = map[string]struct{}{
		"png":  {},
		"jpeg": {},
		"jpg":  {},
	}

	UnsupportedFileTypeErr = errors.New("file extension not supported")
)

// IsSupportedExt pass in the extension, or path, and this will return the extension + true if the
// extension is supported - false if not supported
func IsSupportedFile(path string) (string, bool) {
	ext := strings.ToLower(strings.ReplaceAll(filepath.Ext(path), ".", ""))
	_, ok := supportedTypes[ext]
	return ext, ok
}

// Raw again does the same as other functions, but can be used when getting image data directly from
// a device, such as a webcam stream
func Raw(frame []byte, opts ScaleOpts) (image.Image, error) {
	img, err := jpeg.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	scaled := Image(img, opts)
	return scaled, nil
}

// File does the same thing as Image, but takes a string which should be a valid path to an image file
// it opens it, scales it, and returns the scaled image
func File(imgFile string, opts ScaleOpts) (image.Image, error) {
	inF, err := os.Open(imgFile)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(strings.ReplaceAll(filepath.Ext(imgFile), ".", ""))
	if _, ok := supportedTypes[ext]; !ok {
		return nil, UnsupportedFileTypeErr
	}
	var src image.Image
	if ext == "png" {
		src, err = png.Decode(inF)
	} else {
		src, err = jpeg.Decode(inF)
	}
	// close file, we're done
	inF.Close()
	if err != nil {
		return nil, err
	}
	// we have out image, now we can scale it
	scaled := Image(src, opts)
	return scaled, nil
}

func FileToWindow(imgFile string, opts ScaleOpts) (image.Image, error) {
	inF, err := os.Open(imgFile)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(strings.ReplaceAll(filepath.Ext(imgFile), ".", ""))
	if _, ok := supportedTypes[ext]; !ok {
		return nil, UnsupportedFileTypeErr
	}
	var src image.Image
	if ext == "png" {
		src, err = png.Decode(inF)
	} else {
		src, err = jpeg.Decode(inF)
	}
	inF.Close()
	// determine factor
	if opts.Width != 0 && opts.Height != 0 && opts.Factor != 0 {
		max := src.Bounds().Max
		useFact := false
		if uint(max.Y) > opts.Height {
			hf := float64(opts.Height) / float64(max.Y)
			if hf < opts.Factor {
				opts.Factor = hf
				useFact = true
			}
		}
		if uint(max.X) > opts.Width {
			hf := float64(opts.Width) / float64(max.X)
			if hf < opts.Factor {
				opts.Factor = hf
				useFact = true
			}
		}
		// true size
		if opts.Factor > 1.0 {
			opts.Factor = 1.0
		}
		if useFact {
			opts.Width = 0
			opts.Height = 0
		}
	}
	scaled := Image(src, opts)
	return scaled, nil
}

// Image takes a given image, and returns a scaled version thereof
func Image(src image.Image, opts ScaleOpts) image.Image {
	x, y := getScaledXY(opts, src)
	dst := image.NewRGBA(image.Rect(0, 0, x, y))
	switch opts.Mode {
	case NearestNeighbourScaling:
		draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	case ApproxBilinearScaling:
		draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	case BilinearScaling:
		draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	case CatmullRomScaling:
		draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	}
	return dst
}

// scale the current source image accorind to factor, unless width && height are set, then just use those
func getScaledXY(opts ScaleOpts, src image.Image) (int, int) {
	if opts.Factor == 0 {
		return int(opts.Width), int(opts.Height)
	}
	max := src.Bounds().Max
	x, y := math.Round(float64(max.X)*opts.Factor), math.Round(float64(max.Y)*opts.Factor)
	return int(x), int(y)
}

// String returns the scale mode as string
func (s Mode) String() string {
	m, ok := scaleStr[s]
	if !ok {
		return ""
	}
	return m
}
