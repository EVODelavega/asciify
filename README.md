# ASCIIfy

Convert an image to ASCII character-based string output

## Compile

It's pure go, so just run `go build .` and you're good

## Run

The help output details all the flags:

```
  -A	Print image as ASCII chars (default false)
  -f string
    	Input file
  -h uint
    	The height to resize the image to
  -m string
    	Choose scaling algorithm (fast & low quality to slow but high quality: near [Nearest Neighbour], approx [Approximate Bilinear], bilinear [Bilinear], cat [CatmullRom]) (default "near")
  -n	Make negative of the ASCII output (white <> black)
  -o string
    	Output file - default is output.txt
  -r	Replace output file if exists
  -s float
    	The scaling factor to use instead of width/height float value (default 1)
  -w uint
    	The width to resize the image to
  -c string
    	Save a copy of the scaled image under given file name
```

Check the examples directory for an image that was downscaled, and the ASCII output it generated.

The commands used were 

```bash
asciify -f example/teapot.jpg -o example/output.txt -s 0.1
asciify -f example/teapot.jpg -o example/output_neg.txt -n -s 0.1
```

This scales the image down to 1/10th its original size and outputs the characters in the specified files. To keep the scaled version of the input file, just specify a -c flag value

```bash
asciify -f example/teapot.jpg -s 0.1 -c teapot_resized.jpg
```

This will write the output.txt file to the current directory

These commands use a scaling factor, which preserves the aspect ratio of the original image. If the output looks a bit stretched, which can happen, you can specify a width and height manually to compensate:

```bash
asciify -f example/teapot.jpg -w 200 -h 180 -o example/output_wh.txt
asciify -f example/teapot.jpg -w 200 -h 180 -o example/output_whn.txt -n
```

## Credit

The image used in the example directory is a royalty-free image from [The Graphics Fairy](https://thegraphicsfairy.com)
