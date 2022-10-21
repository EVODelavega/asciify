package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/EVODelavega/asciify/convert"
	"github.com/EVODelavega/asciify/scale"
	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type Args struct {
	scale.ScaleOpts
	Cam              string
	X, Y             uint // input stream resolution
	negative, invert bool
}

func main() {
	ctx, cfunc := context.WithCancel(context.Background())
	defer cfunc()
	sCh := make(chan os.Signal, 1)
	done := make(chan struct{})

	go func() {
		<-sCh
		cfunc()
		close(done)
		close(sCh)
	}()
	signal.Notify(
		sCh,
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
		syscall.SIGKILL,
	)
	args := Args{}
	// just use the fastest scaling
	args.Mode = scale.NearestNeighbourScaling
	flag.UintVar(&args.Width, "w", 0, "ASCII width (number of columns)")
	flag.UintVar(&args.Height, "h", 0, "ASCII height (number of rows)")
	flag.Float64Var(&args.Factor, "s", 1.0, "The scaling factor to use instead of width/height float value")
	flag.StringVar(&args.Cam, "d", "/dev/video0", "Input device")
	flag.BoolVar(&args.negative, "n", false, "Show negative image (black <> white)")
	flag.BoolVar(&args.invert, "i", true, "Invert image (mirror output)")
	flag.UintVar(&args.X, "x", 640, "Input camera resolution (width/X)")
	flag.UintVar(&args.Y, "y", 480, "Input camera resolution (height/Y)")
	// cmd := exec.Command("clear")
	// cmd.Stdout = os.Stdout
	flag.Parse()
	// wipe factor if width and height were set
	if args.Width != 0 && args.Height != 0 {
		args.Factor = 0
	}
	camera, err := device.Open(
		args.Cam,
		device.WithPixFormat(v4l2.PixFormat{
			PixelFormat: v4l2.PixelFmtMJPEG,
			Width:       uint32(args.X),
			Height:      uint32(args.Y),
		}),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer camera.Close()
	if err := camera.Start(ctx); err != nil {
		log.Fatalf("camera start: %s", err)
	}
	for frame := range camera.GetOutput() {
		img, err := scale.Raw(frame, args.ScaleOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ASCIIStr := convert.ImgToASCII(img, args.negative, args.invert)
		clear()
		fmt.Printf("\n%s\n", ASCIIStr)
	}
	<-done
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
