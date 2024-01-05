package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/go-getoptions"
	"golang.org/x/image/draw"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))

	resize := opt.NewCommand("resize", "Resize an image")
	resize.SetCommandFn(ResizeRun)
	resize.HelpSynopsisArg("<image_file>", "Base Image file")
	resize.HelpSynopsisArg("<size>", "Size of the image, only one necessary since it keeps aspect ratio")
	resize.String("output", "output.png", opt.Alias("o"), opt.Description("Output file name"))

	set := opt.NewCommand("set", "Create a set of images with different sizes")
	set.SetCommandFn(ResizeRun)
	set.HelpSynopsisArg("<image_file>", "Base Image file, used for 3x size")
	set.String("output", "output.png", opt.Alias("o"), opt.Description("Output file name"))
	set.Int("size", 0, opt.Description("Size of the 3x image, only one necessary since it keeps aspect ratio"))

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func ResizeRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")
	output := opt.Value("output").(string)

	imgFile, args, err := opt.GetRequiredArg(args)
	if err != nil {
		return err
	}
	size, _, err := opt.GetRequiredArgInt(args)
	if err != nil {
		return err
	}

	err = resizeImg(ctx, imgFile, output, size)
	if err != nil {
		return err
	}

	return nil
}

func SetResizeRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")
	output := opt.Value("output").(string)

	imgFile, args, err := opt.GetRequiredArg(args)
	if err != nil {
		return err
	}
	size, _, err := opt.GetRequiredArgInt(args)
	if err != nil {
		return err
	}

	err = resizeImg(ctx, imgFile, output, size)
	if err != nil {
		return err
	}

	return nil
}

func resizeImg(ctx context.Context, imgFile, outputFile string, size int) error {
	img, err := os.Open(imgFile)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer img.Close()

	output, _ := os.Create(outputFile)
	defer output.Close()

	var imgSrc image.Image
	if strings.HasSuffix(imgFile, ".png") {
		imgSrc, err = png.Decode(img)
		if err != nil {
			return fmt.Errorf("failed to decode png image: %w", err)
		}
	}
	if strings.HasSuffix(imgFile, ".jpg") || strings.HasSuffix(imgFile, ".jpeg") {
		imgSrc, err = jpeg.Decode(img)
		if err != nil {
			return fmt.Errorf("failed to decode jpeg image: %w", err)
		}
	}
	max := imgSrc.Bounds().Max.X
	if imgSrc.Bounds().Max.Y > max {
		max = imgSrc.Bounds().Max.Y
	}
	ratio := float64(max) / float64(size)
	Logger.Printf("Ratio: %f", ratio)
	dstX := int(float64(imgSrc.Bounds().Max.X) / ratio)
	dstY := int(float64(imgSrc.Bounds().Max.Y) / ratio)
	Logger.Printf("Resizing image from %dx%d to %dx%d", imgSrc.Bounds().Max.X, imgSrc.Bounds().Max.Y, dstX, dstY)

	dst := image.NewRGBA(image.Rect(0, 0, dstX, dstY))

	draw.CatmullRom.Scale(dst, dst.Bounds(), imgSrc, imgSrc.Bounds(), draw.Over, nil)
	err = png.Encode(output, dst)
	if err != nil {
		return fmt.Errorf("failed to encode png image: %w", err)
	}
	return nil
}
