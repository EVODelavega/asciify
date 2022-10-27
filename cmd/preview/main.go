package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/EVODelavega/asciify/convert"
	"github.com/EVODelavega/asciify/scale"
)

var (
	ErrInvalidInputFormat   = errors.New("unsupported input type")
	ErrMissingInputFile     = errors.New("input file not specified or missing")
	ErrInvalidScalingMethod = errors.New("specified scaling mode not supported")

	// flag values map onto constants
	scaleModes = map[string]scale.Mode{
		"near":     scale.NearestNeighbourScaling,
		"approx":   scale.ApproxBilinearScaling,
		"bilinear": scale.BilinearScaling,
		"cat":      scale.CatmullRomScaling,
	}
)

// Conf just groups the flags together
type Conf struct {
	scale.ScaleOpts
	in    string
	force bool
}

func (c *Conf) validate() error {
	// force no scaling factor if window width/height are set
	if c.Width != 0 && c.Height != 0 {
		c.Factor = 1.0
		if c.force {
			c.Factor = 0
		}
	}
	if c.in == "" || !fileExists(c.in) {
		return ErrMissingInputFile
	}
	if _, ok := scale.IsSupportedFile(c.in); !ok {
		return ErrInvalidInputFormat
	}
	return nil
}

func main() {
	var conf Conf
	var scaleFlag, scaleDoc string
	flags := make([]string, 0, len(scale.OrderLHQ))
	scaleFlag = scaleModeFlagStr(scale.OrderLHQ[0])
	for _, s := range scale.OrderLHQ {
		flags = append(flags, fmt.Sprintf("%s [%s]", scaleModeFlagStr(s), s.String()))
	}
	scaleDoc = fmt.Sprintf("Choose scaling algorithm (fast & low quality to slow but high quality: %s)", strings.Join(flags, ", "))
	flag.UintVar(&conf.Width, "w", 0, "Max width - scales image (if required) to fit max width. recalculates -s flag")
	flag.UintVar(&conf.Height, "h", 0, "Max height - scales image (if required) to fit max height. recalculates -s flag")
	flag.Float64Var(&conf.Factor, "s", 1.0, "The scaling factor to use instead of width/height float value")
	flag.StringVar(&conf.in, "f", "", "Input file")
	flag.StringVar(&scaleFlag, "m", scaleFlag, scaleDoc)
	flag.BoolVar(&conf.force, "S", false, "Force width and height to be used as absolute ratio - Ignore s flag")
	flag.Parse()
	if err := conf.validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	smode, err := scaleModeFromFalgStr(scaleFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	conf.Mode = smode
	scaled, err := scale.FileToWindow(conf.in, conf.ScaleOpts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	strImg := convert.ImgToPreview(scaled)
	fmt.Println(strImg)
}

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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
