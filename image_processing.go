/*
That's basically just the first idea, I still have to understand some of the code later because I haven't gone through it all properly yet
also see about other file formats...
and maybe other transformation
*/

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"math"
)

func grayscale(loadedImage image.Image, output_img *image.NRGBA, width int, height int) {
	//loops through the whole array (all the pixels of the image) and calculates the average value of r, g, b
	//by setting the average value as the value for each color we have a grayscale image
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			colors := loadedImage.At(i, j)
			r, g, b, a := colors.RGBA()
			r = uint32(map_int(int(r), 65535, 255))
			g = uint32(map_int(int(g), 65535, 255))
			b = uint32(map_int(int(b), 65535, 255))
			a = uint32(map_int(int(a), 65535, 255))
			avg := (r + g + b) / 3

			output_img.Set(i, j, color.NRGBA{
				R: uint8(avg),
				G: uint8(avg),
				B: uint8(avg),
				A: uint8(a),
			})
		}
	}
}

func gaussian_blur(loadedImage image.Image, output_img *image.NRGBA, width int, height int, radius int) {
	sigma := math.Max(float64((radius / 2)), 1)
	pi := math.Pi
	kernel_width := (2 * radius) + 1
	conv_kernel := make([][]float64, kernel_width)
	for i := range conv_kernel {
		conv_kernel[i] = make([]float64, kernel_width)
	}

	for i := -radius; i < radius; i++ {
		for j := -radius; j < radius; j++ {
			divider := 2 * pi * sigma * sigma
			exponential_numerator := float64(-(i * i + j + j))
			exponential_denominator := 2 * sigma * sigma
			conv_kernel[i + radius][j + radius] = float64((math.Exp(exponential_numerator/exponential_denominator)/divider))
		}
	}
	fmt.Println(conv_kernel)
}

func main() {
	// Read image from file that already exists
	existingImageFile, err := os.Open("pic.png")
	if err != nil {
		log.Fatal(err)
	}

	defer existingImageFile.Close()

	loadedImage, err := png.Decode(existingImageFile)
	if err != nil {
		log.Fatal(err)
	}
	width := loadedImage.Bounds().Max.X
	height := loadedImage.Bounds().Max.Y
	radius := 3

	output_img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// grayscale(loadedImage, output_img, width, height)
	gaussian_blur(loadedImage, output_img, width, height, radius)

	dir_path := "output"
	err = os.Mkdir(dir_path, 0755)
	output_path := "./output/image.png"
	f, err := os.Create(output_path)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, output_img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Image generated at", output_path)
}

func map_int(to_map int, max_from int, max_to int) int {
	return (to_map * max_to) / max_from
}
