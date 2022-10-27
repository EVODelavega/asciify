package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/EVODelavega/asciify/convert"
	"github.com/EVODelavega/asciify/scale"
)

// Config is basically all the flags so we can check/validate them easily
type Config struct {
	scale.ScaleOpts
	in, out    string
	overwrite  bool
	printASCII bool
	reverse    bool
	saveScaled string
	colour     bool

	// not flags, but avoid doing the getting extensions a second time
	inExt, outExt string
}

var (
	ErrInvalidDimensions    = errors.New("need valid width/height or factor")
	ErrInvalidInputFormat   = errors.New("unsupported input type")
	ErrMissingInputFile     = errors.New("input file not specified or missing")
	ErrInvalidScalingMethod = errors.New("specified scaling mode not supported")
	ErrOutputFileExists     = errors.New("output file already exists")

	// flag values map onto constants
	scaleModes = map[string]scale.Mode{
		"near":     scale.NearestNeighbourScaling,
		"approx":   scale.ApproxBilinearScaling,
		"bilinear": scale.BilinearScaling,
		"cat":      scale.CatmullRomScaling,
	}
)

func scaleModeFromFalgStr(fs string) (scale.Mode, error) {
	m, ok := scaleModes[fs]
	if !ok {
		return m, ErrInvalidScalingMethod
	}
	return m, nil
}

func scaleModeFlagStr(s scale.Mode) string {
	for k, v := range scaleModes {
		if v == s {
			return k
		}
	}
	return ""
}

// Validate makes sure the config makes sense - mode is handled in main function though
func (c *Config) Validate() error {
	if c.Width == 0 && c.Height == 0 {
		// we need a valid factor
		if c.Factor <= 0 {
			return ErrInvalidDimensions
		}
	}
	if c.Factor == 0 && (c.Width == 0 || c.Height == 0) {
		return ErrInvalidDimensions
	}
	// clear whichever dimension values we won't use
	if c.Width != 0 && c.Height != 0 {
		c.Factor = 0
	} else {
		c.Width, c.Height = 0, 0
	}
	if c.in == "" || !fileExists(c.in) {
		return ErrMissingInputFile
	}
	ext, ok := scale.IsSupportedFile(c.in)
	if !ok {
		return ErrInvalidInputFormat
	}
	c.inExt = ext
	if c.out == "" {
		c.out = "output.txt"
	}
	if !c.overwrite && fileExists(c.out) {
		return ErrOutputFileExists
	}
	if len(c.saveScaled) > 0 {
		// we're going to assume this is an accepted format, it'll write the default (JPEG) anyway
		c.outExt = strings.ToLower(strings.ReplaceAll(filepath.Ext(c.saveScaled), ".", ""))
	}
	return nil
}

func main() {
	conf := Config{}
	var scaleFlag, scaleDoc string
	flags := make([]string, 0, len(scale.OrderLHQ))
	scaleFlag = scaleModeFlagStr(scale.OrderLHQ[0])
	for _, s := range scale.OrderLHQ {
		flags = append(flags, fmt.Sprintf("%s [%s]", scaleModeFlagStr(s), s.String()))
	}
	scaleDoc = fmt.Sprintf("Choose scaling algorithm (fast & low quality to slow but high quality: %s)", strings.Join(flags, ", "))
	flag.UintVar(&conf.Width, "w", 0, "The width to resize the image to")
	flag.UintVar(&conf.Height, "h", 0, "The height to resize the image to")
	flag.Float64Var(&conf.Factor, "s", 1.0, "The scaling factor to use instead of width/height float value")
	flag.StringVar(&conf.in, "f", "", "Input file")
	flag.StringVar(&conf.out, "o", "", "Output file - default is output.txt")
	flag.StringVar(&scaleFlag, "m", scaleFlag, scaleDoc)
	flag.BoolVar(&conf.overwrite, "r", false, "ReplaceAll output file if exists")
	flag.BoolVar(&conf.printASCII, "A", false, "Print image as ASCII chars")
	flag.BoolVar(&conf.reverse, "n", false, "Make negative of the ASCII output (white <> black)")
	flag.BoolVar(&conf.colour, "C", false, "Show image in colour")
	flag.StringVar(&conf.saveScaled, "c", "", "Save a copy of the scaled image under given file name")

	// get the args
	flag.Parse()
	smode, err := scaleModeFromFalgStr(scaleFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	conf.Mode = smode
	if err := conf.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// valid options, let's get started:
	scaled, err := scale.File(conf.in, conf.ScaleOpts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var strImg string
	// create scaled image string
	if conf.colour {
		strImg = convert.ImgToASCIIColoured(scaled, conf.reverse, false)
	} else {
		strImg = convert.ImgToASCII(scaled, conf.reverse, false)
	}
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
