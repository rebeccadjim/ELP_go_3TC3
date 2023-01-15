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
	"math"
	"os"
	"strconv"
)

func grayscale(loadedImage image.Image, output_img *image.NRGBA, width int, height int) {
	//loops through the whole array (all the pixels of the image) and calculates the average value of r, g, b
	//by setting the average value as the value for each color we have a grayscale image
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			colors := loadedImage.At(i, j)
			red_input, green_input, blue_input, a_input := colors.RGBA()
			red_input = uint32(map_int(int(red_input), 65535, 255))
			green_input = uint32(map_int(int(green_input), 65535, 255))
			blue_input = uint32(map_int(int(blue_input), 65535, 255))
			a_input = uint32(map_int(int(a_input), 65535, 255))
			avg := (red_input + green_input + blue_input) / 3

			output_img.Set(i, j, color.NRGBA{
				R: uint8(avg),
				G: uint8(avg),
				B: uint8(avg),
				A: uint8(a_input),
			})
		}
	}
}

func gaussian_blur(loadedImage image.Image, output_img *image.NRGBA, width int, height int, radius int) {
	sigma := math.Max(float64((radius / 2)), 1)
	pi := math.Pi
	kernel_width := (2 * radius) + 1
	conv_kernel := make([][]float64, kernel_width)
	kernel_sum := float64(0)
	for i := range conv_kernel {
		conv_kernel[i] = make([]float64, kernel_width)
	}

	for i := -radius; i < radius; i++ {
		for j := -radius; j < radius; j++ {
			divider := 2 * pi * sigma * sigma
			exponential_numerator := float64(-(i*i + j + j))
			exponential_denominator := 2 * sigma * sigma
			conv_kernel[i+radius][j+radius] = float64((math.Exp(exponential_numerator/exponential_denominator) / divider))
			kernel_sum += conv_kernel[i+radius][j+radius]
		}
	}
	for i := 0; i < kernel_width; i++ {
		for j := 0; j < kernel_width; j++ {
			conv_kernel[i][j] /= kernel_sum
		}
	}

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			var red_output float64
			var green_output float64
			var blue_output float64
			var a_output float64

			for x := -radius; x < radius; x++ {
				for y := -radius; y < radius; y++ {
					gaussian_coefficient := conv_kernel[x+radius][y+radius]
					x_substitue := x
					y_substitue := y
					for i+x_substitue < 0 {
						x_substitue++
					}
					for j+y_substitue < 0 {
						y_substitue++
					}
					for i+x_substitue >= width {
						x_substitue--
					}
					for j+y_substitue >= height {
						y_substitue--
					}

					image_colors := loadedImage.At(i+x_substitue, j+y_substitue)
					red_input, green_input, blue_input, a_input := image_colors.RGBA()
					red_input = uint32(map_int(int(red_input), 65535, 255))
					green_input = uint32(map_int(int(green_input), 65535, 255))
					blue_input = uint32(map_int(int(blue_input), 65535, 255))
					a_input = uint32(map_int(int(a_input), 65535, 255))

					red_output += float64(red_input) * float64(gaussian_coefficient)
					green_output += float64(green_input) * float64(gaussian_coefficient)
					blue_output += float64(blue_input) * float64(gaussian_coefficient)
					a_output = float64(a_input)
				}
			}
			output_img.Set(i, j, color.NRGBA{
				R: uint8(red_output),
				G: uint8(green_output),
				B: uint8(blue_output),
				A: uint8(a_output),
			})
		}
	}
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
	var gaussian_radius int

	output_img := image.NewNRGBA(image.Rect(0, 0, width, height))

	arguments := os.Args[1:]
	if len(arguments) == 0 {
		fmt.Println("Provide at least on argument: the effect desired")
		os.Exit(1)
	}

	if arguments[0] == "gaussian_blur" {
		if len(arguments) < 2 {
			fmt.Println("Provide a valid radius for the gaussian blur")
			fmt.Println("The radius determines the strength of the blur. The bigger the radius, the stronger the blur")
			os.Exit(1)
		}
		gaussian_radius, err = strconv.Atoi(arguments[1])
		if err != nil {
			fmt.Println("Provide a valid radius for the gaussian blur")
			fmt.Println("The radius determines the strength of the blur. The bigger the radius, the stronger the blur")
			os.Exit(1)
		}
		gaussian_blur(loadedImage, output_img, width, height, gaussian_radius)
	} else if arguments[0] == "grayscale" {
		grayscale(loadedImage, output_img, width, height)
	}

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
