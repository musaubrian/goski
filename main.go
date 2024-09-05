package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KononK/resize"
	"golang.org/x/term"
)

func main() {
	var (
		imgPath      string
		useColor     bool
		useAutoScale bool
		remote       bool
	)

	flag.StringVar(&imgPath, "img", "", "Path to image file")
	flag.BoolVar(&useColor, "c", false, "Output colored ascii art")
	flag.BoolVar(&useAutoScale, "s", true, "Auto scale the output to fit in your term")
	flag.BoolVar(&remote, "r", false, "Allows you to pass in a url to the image")
	flag.Parse()

	if len(imgPath) == 0 {
		flag.Usage()
		return
	}

	if strings.Contains(imgPath, "http") && !remote {
		slog.Error("To use a remote resource, please use the `-r` flag")
		return
	} else if remote && !strings.Contains(imgPath, "http") {
		filepath, err := getRemoteResource(imgPath)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		imgPath = filepath
	}

	f, err := os.Open(imgPath)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open:\n%s", err.Error()))
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to decode: %s", err.Error()))
		return
	}

	if useAutoScale {
		termWidth, termHeight := getTerminalSize()
		nw, nh := autoScale(img.Bounds().Dx(), img.Bounds().Dy(), termWidth, termHeight)
		img = resize.Resize(uint(nw), uint(nh), img, resize.MitchellNetravali)
	}

	asciiChars := " .,:;i1tfLCG08@"

	font := strings.Split(asciiChars, "")
	if useColor {
		asciiOutput := coloredAsciiOutput(img, font)
		fmt.Print(asciiOutput)
		return
	}

	asciiOutput := grayScaledAscii(img, font)
	fmt.Print(asciiOutput)
}

func grayScaledAscii(img image.Image, font []string) string {
	var asciiImgChars strings.Builder

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			char := int(c.Y) * (len(font) - 1) / 255
			asciiImgChars.WriteString(font[char])

		}
		asciiImgChars.WriteString("\n")
	}
	return asciiImgChars.String()
}

func coloredAsciiOutput(img image.Image, font []string) string {
	var coloredAscii strings.Builder
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			// Normalize the RGB values to 0-255 range
			r = r >> 8
			g = g >> 8
			b = b >> 8
			// average intensity
			avg := (r + g + b) / 3
			char := int(avg) * (len(font) - 1) / 255

			coloredAscii.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", r, g, b, font[char]))
		}
		coloredAscii.WriteString(fmt.Sprint("\n\x1b[0m"))
	}
	return coloredAscii.String()
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

func getRemoteResource(url string) (string, error) {
	errCh := make(chan error)
	spinnerCh := make(chan bool, 1)
	bodyCh := make(chan []byte)

	go spinner(spinnerCh)
	go func() {
		defer func() {
			spinnerCh <- false

			close(errCh)
			close(bodyCh)
		}()

		spinnerCh <- true
		r, err := http.Get(url)
		if err != nil {
			errCh <- err
			return
		}
		defer r.Body.Close()

		if r.StatusCode != 200 {
			errCh <- fmt.Errorf("Could not find resource")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			errCh <- err
			return
		}
		bodyCh <- body
	}()

	file := "/tmp/" + generateName()
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}

	defer f.Close()

	select {
	case body := <-bodyCh:
		_, err = f.Write(body)
		if err != nil {
			return "", err
		}
	case err := <-errCh:
		return "", err

	}

	return file, <-errCh
}

func spinner(spinnerCh <-chan bool) {
	charset := "⢿⣻⣽⣾⣷⣯⣟⡿"
	chars := strings.Split(charset, "")

	for i := 0; ; i++ {
		select {
		case run := <-spinnerCh:
			if !run {
				return
			}
		default:
			fmt.Printf("%s\r\b", chars[i%len(chars)])
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func generateName() string {
	wl := 3
	cs := "bcdfghjklmnpqrstvwxzy"
	vs := "aeiou"
	var result string

	for i := 0; i < wl; i++ {
		result += string(cs[rand.IntN(len(cs))])
		result += string(vs[rand.IntN(len(vs))])
	}
	return "img.goski_" + result
}
