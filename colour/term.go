// Package colour takes a color.Color value, and maps it on to the corresponding terminal value
package colour

import (
	"fmt"
	"image/color"
	"strconv"
)

const (
	// ResetColour is the terminal code to reset/turn off colour output
	ResetColour = "\033[0m"

	trueColourF = "\033[38;2;%d;%d;%dm"
)

// Colour256 contains values for all three (RGB) channels (no alpha)
// as values ranging from 0 to 255 (0x00 - 0xFF)
type Colour256 struct {
	R, G, B uint8
}

// FromColor returns nil for transparent pixels
func FromColor(c color.Color) *Colour256 {
	r, g, b, a := c.RGBA()
	if a == 0 {
		return nil
	}
	return &Colour256{
		R: uint8(r / 256), // yields values between 0 and 255
		G: uint8(g / 256),
		B: uint8(b / 256),
	}
}

// FromHex returns a Colour256 from a given hex string
func FromHex(hex string) (*Colour256, error) {
	val, err := strconv.ParseUint(string(hex), 16, 32)
	if err != nil {
		return nil, err
	}
	return &Colour256{
		R: uint8(val >> 16),
		G: uint8((val >> 8) & 0xFF),
		B: uint8(val & 0xFF),
	}, nil
}

// TrueEsc returns true-colour escape code
func (c Colour256) TrueEsc() string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", c.R, c.G, c.B)
}

// Hex returns 256 colour as a hex string
func (c Colour256) Hex() string {
	return fmt.Sprintf("0x%02x%02x%02x", c.R, c.G, c.B)
}

// UInt get the hex value as a single uint value (in 0x000000 to 0xffffff range)
func (c Colour256) UInt() uint {
	r, g, b := uint(c.R), uint(c.G), uint(c.B)
	return (r << 16) + (g << 8) + b
}
