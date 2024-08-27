package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"os"
	"strings"

	"github.com/KononK/resize"
	"golang.org/x/term"
)

func main() {
	var imgPath string
	var useColor bool
	var useAutoScale bool

	flag.StringVar(&imgPath, "img", "", "Path to image file")
	flag.BoolVar(&useColor, "color", false, "Output colored ascii art")
	flag.BoolVar(&useAutoScale, "scale", false, "Auto scale the output to fit in your term")
	flag.Parse()

	if len(imgPath) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(imgPath)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open:\n%s", err.Error()))
		os.Exit(1)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to decode: %s", err.Error()))
		os.Exit(1)
	}

	if useAutoScale {
		termWidth, termHeight := getTerminalSize()
		nw, nh := autoScale(img.Bounds().Dx(), img.Bounds().Dy(), termWidth, termHeight)
		img = resize.Resize(uint(nw), uint(nh), img, resize.MitchellNetravali)
	}

	asciiChars := " .,:;i1tfLCG08@"
	font := strings.Split(asciiChars, "")
	if useColor {
		coloredAsciiOutput(img, font)
	} else {
		grayScaledAscii(img, font)
	}

}
func grayScaledAscii(img image.Image, font []string) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			char := int(c.Y) * (len(font) - 1) / 255
			fmt.Print(font[char])
		}
		fmt.Println()
	}
}

func coloredAsciiOutput(img image.Image, font []string) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			// Normalize the RGB values to 0-255 range
			r = r >> 8
			g = g >> 8
			b = b >> 8
			// Calculate average intensity
			avg := (r + g + b) / 3
			char := int(avg) * (len(font) - 1) / 255

			fmt.Printf("\x1b[38;2;%d;%d;%dm%s", r, g, b, font[char])
		}
		fmt.Println("\x1b[0m")
	}
}

func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return width, height
}

func autoScale(imgWidth, imgHeight, termWidth, termHeight int) (int, int) {
	// 0.95 is so that the image doesn't fill the entire height of the term
	targetHeight := int(float64(termHeight) * 0.95)

	ratio := float64(targetHeight) / float64(imgHeight)
	newWidth := int(float64(imgWidth) * ratio * 2.5)

	if newWidth > termWidth {
		ratio = float64(termWidth) / float64(imgWidth)
		targetHeight = int(float64(imgHeight) * ratio)
		newWidth = termWidth
	}

	return newWidth, targetHeight
}
