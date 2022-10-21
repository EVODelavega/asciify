package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

type ScaleMode uint32

// Config is basically all the flags so we can check/validate them easily
type Config struct {
	w, h       uint
	fact       float64
	in, out    string
	mode       ScaleMode
	overwrite  bool
	printASCII bool
	reverse    bool
	saveScaled string

	// not flags, but avoid doing the getting extensions a second time
	inExt, outExt string
}

type PixelChar struct {
	x, y int
	char rune
}

const (
	NearestNeighbourScaling ScaleMode = iota
	ApproxBilinearScaling
	BilinearScaling
	CatmullRomScaling

	// where this value comes from is explained at ASCIIChars
	// working method step
	CharStep = 34 * 257
)

var supportedTypes = map[string]struct{}{
	"jpg":  {},
	"jpeg": {},
	"png":  {},
}

var (
	InvalidDimensionsErr    = errors.New("need valid width/height or factor")
	InvalidInputFormatErr   = errors.New("unsupported input type")
	MissingInputFileErr     = errors.New("input file not specified or missing")
	InvalidScalingMethodErr = errors.New("specified scaling mode not supported")
	OutputFileExistsErr     = errors.New("output file already exists")

	// 30 chars -> with RGBA (255 * 4), each character is used for 34 shades
	// color.Color represents RGBA values as 0-0xFFFF (65,535), so 34 * 257
	ASCIIChars = []rune("Ñ@#W$9876543210?!abc;:+=-,._ ")

	// flag values map onto constants
	scaleModes = map[string]ScaleMode{
		"near":     NearestNeighbourScaling,
		"approx":   ApproxBilinearScaling,
		"bilinear": BilinearScaling,
		"cat":      CatmullRomScaling,
	}

	// for human-readible representation
	scaleStr = map[ScaleMode]string{
		NearestNeighbourScaling: "Nearest Neighbour",
		ApproxBilinearScaling:   "Approximate Bilinear",
		BilinearScaling:         "Bilinear",
		CatmullRomScaling:       "CatmullRom",
	}

	// order of scaling fast -> high quality
	scaleOrder = []ScaleMode{
		NearestNeighbourScaling,
		ApproxBilinearScaling,
		BilinearScaling,
		CatmullRomScaling,
	}
)

func (s *ScaleMode) FromFlag(fs string) error {
	m, ok := scaleModes[fs]
	if !ok {
		return InvalidScalingMethodErr
	}
	*s = m
	return nil
}

func (s ScaleMode) String() string {
	str, ok := scaleStr[s]
	if !ok {
		return ""
	}
	return str
}

func (s ScaleMode) FlagStr() string {
	for k, v := range scaleModes {
		if v == s {
			return k
		}
	}
	return ""
}

// Validate makes sure the config makes sense - mode is handled in main function though
func (c *Config) Validate() error {
	if c.w == 0 && c.h == 0 {
		// we need a valid factor
		if c.fact <= 0 {
			return InvalidDimensionsErr
		}
	}
	if c.fact == 0 && (c.w == 0 || c.h == 0) {
		return InvalidDimensionsErr
	}
	// clear whichever dimension values we won't use
	if c.w != 0 && c.h != 0 {
		c.fact = 0
	} else {
		c.w, c.h = 0, 0
	}
	if c.in == "" || !fileExists(c.in) {
		return MissingInputFileErr
	}
	ext := strings.ToLower(strings.ReplaceAll(filepath.Ext(c.in), ".", ""))
	if _, ok := supportedTypes[ext]; !ok {
		return InvalidInputFormatErr
	}
	c.inExt = ext
	if c.out == "" {
		c.out = "output.txt"
	}
	if !c.overwrite && fileExists(c.out) {
		return OutputFileExistsErr
	}
	if len(c.saveScaled) > 0 {
		ext := strings.ToLower(strings.ReplaceAll(filepath.Ext(c.saveScaled), ".", ""))
		if _, ok := supportedTypes[ext]; !ok {
			return InvalidInputFormatErr
		}
		c.outExt = ext
	}
	return nil
}

func main() {
	conf := Config{}
	var scaleFlag, scaleDoc string
	flags := make([]string, 0, len(scaleOrder))
	scaleFlag = scaleOrder[0].FlagStr()
	for _, s := range scaleOrder {
		flags = append(flags, fmt.Sprintf("%s [%s]", s.FlagStr(), s.String()))
	}
	scaleDoc = fmt.Sprintf("Choose scaling algorithm (fast & low quality to slow but high quality: %s)", strings.Join(flags, ", "))
	flag.UintVar(&conf.w, "w", 0, "The width to resize the image to")
	flag.UintVar(&conf.h, "h", 0, "The height to resize the image to")
	flag.Float64Var(&conf.fact, "s", 1.0, "The scaling factor to use instead of width/height float value")
	flag.StringVar(&conf.in, "f", "", "Input file")
	flag.StringVar(&conf.out, "o", "", "Output file - default is output.txt")
	flag.StringVar(&scaleFlag, "m", scaleFlag, scaleDoc)
	flag.BoolVar(&conf.overwrite, "r", false, "ReplaceAll output file if exists")
	flag.BoolVar(&conf.printASCII, "A", false, "Print image as ASCII chars")
	flag.BoolVar(&conf.reverse, "n", false, "Make negative of the ASCII output (white <> black)")
	flag.StringVar(&conf.saveScaled, "c", "", "Save a copy of the scaled image under given file name")

	// get the args
	flag.Parse()
	if err := conf.mode.FromFlag(scaleFlag); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := conf.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// valid options, let's get started:
	scaled, err := scaleImg(conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// create scaled image string
	strImg := imgToASCII(conf, scaled)
	// first, write the scaled copy
	if err := writeOut(conf, strImg); err != nil {
		fmt.Println(err)
	}
	if len(conf.saveScaled) > 0 {
		if err := saveScaledImg(conf, scaled); err != nil {
			fmt.Println(err)
		}
	}
	if !conf.printASCII {
		return
	}
	fmt.Println(strImg)
}

func imgToASCII(c Config, scaled image.Image) string {
	// now we can work on the scaled image to ASCII part
	max := scaled.Bounds().Max
	wg := sync.WaitGroup{}
	wg.Add(max.Y)
	done := make(chan struct{})             // the routine that will populate the slice  will let us know when it's done with this
	ch := make(chan PixelChar, max.Y+max.X) // buffer enough for first pixels of each row + 1 column
	matrix := make([][]rune, max.Y)         // matrix[height][width]
	// start waiting for data
	go func() {
		for pc := range ch {
			matrix[pc.y][pc.x] = pc.char
		}
		close(done)
	}()
	for y := 0; y < max.Y; y++ {
		matrix[y] = make([]rune, max.X) // initialise each column
		go convertRow(&wg, ch, scaled, y, c.reverse)
	}
	wg.Wait()
	close(ch)
	<-done
	// OK, our slice is populated, convert to string:

	chunks := make([]string, 0, len(matrix))
	for _, r := range matrix {
		// trim trailing spaces
		chunks = append(chunks, strings.TrimRight(string(r), " "))
	}
	return strings.Join(chunks, "\n")
}

func convertRow(wg *sync.WaitGroup, ch chan<- PixelChar, img image.Image, y int, reverse bool) {
	max := img.Bounds().Max.X
	for x := 0; x < max; x++ {
		// alpha is already applied, so we can just ignore it
		i := 0
		r, g, b, a := img.At(x, y).RGBA()
		if a == 0 {
			// alpha on max, space character
			i = len(ASCIIChars) - 1
		} else {
			i = int(r+g+b) / CharStep
			if !reverse {
				i = len(ASCIIChars) - i
			}
		}
		ch <- PixelChar{
			char: ASCIIChars[i%len(ASCIIChars)],
			x:    x,
			y:    y,
		}
	}
	wg.Done()
}

// scaleImg
func scaleImg(c Config) (image.Image, error) {
	in, err := getInput(c)
	if err != nil {
		return nil, err
	}
	out := createOut(c, in)
	return out, nil
}

func getInput(c Config) (image.Image, error) {
	inF, err := os.Open(c.in)
	if err != nil {
		return nil, err
	}
	var src image.Image
	if c.inExt == "png" {
		src, err = png.Decode(inF)
	} else {
		src, err = jpeg.Decode(inF)
	}
	// close file, we're done
	inF.Close()
	if err != nil {
		return nil, err
	}
	return src, nil
}

func writeOut(c Config, ascii string) error {
	if c.overwrite && fileExists(c.out) {
		os.Remove(c.out)
	}
	output, err := os.Create(c.out)
	if err != nil {
		return err
	}
	_, err = output.WriteString(ascii)
	output.Close()
	return err
}

func saveScaledImg(c Config, scaled image.Image) error {
	if c.overwrite && fileExists(c.saveScaled) {
		os.Remove(c.saveScaled)
	}
	output, err := os.Create(c.saveScaled)
	if err != nil {
		return err
	}
	if c.outExt == "png" {
		png.Encode(output, scaled)
	} else {
		jpeg.Encode(output, scaled, &jpeg.Options{
			Quality: 100, // max quality
		})
	}
	output.Close()
	return nil
}

func getScaledXY(c Config, src image.Image) (int, int) {
	if c.fact == 0 {
		return int(c.w), int(c.h)
	}
	max := src.Bounds().Max
	x, y := math.Round(float64(max.X)*c.fact), math.Round(float64(max.Y)*c.fact)
	return int(x), int(y)
}

func createOut(c Config, src image.Image) image.Image {
	x, y := getScaledXY(c, src)
	// use src.ColorModel() and convert if needed
	dst := image.NewRGBA(image.Rect(0, 0, x, y))
	// with sanitised inputs, we shouldn't need a default case here
	switch c.mode {
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
