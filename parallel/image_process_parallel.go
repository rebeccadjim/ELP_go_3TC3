package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

var running bool
var numGoroutines int

type matrix_elem struct {
	index_i int
	index_j int
}

type Pixel struct {
	index_i int
	index_j int
	pigment color.NRGBA
}

func grayscale_worker(jobChan chan matrix_elem, resultChan chan Pixel, loadedImage image.Image) {
	//loops through the whole array (all the pixels of the image) and calculates the average value of r, g, b
	//by setting the average value as the value for each color we have a grayscale image
	for running {
		elem := <-jobChan

		colors := loadedImage.At(elem.index_i, elem.index_j)
		red_input, green_input, blue_input, a_input := colors.RGBA()
		red_input = uint32(map_int(int(red_input), 65535, 255))
		green_input = uint32(map_int(int(green_input), 65535, 255))
		blue_input = uint32(map_int(int(blue_input), 65535, 255))
		a_input = uint32(map_int(int(a_input), 65535, 255))
		avg := (red_input + green_input + blue_input) / 3

		color := color.NRGBA{
			R: uint8(avg),
			G: uint8(avg),
			B: uint8(avg),
			A: uint8(a_input),
		}
		pixel := Pixel{
			index_i: elem.index_i,
			index_j: elem.index_j,
			pigment: color,
		}

		resultChan <- pixel
	}
}

func create_convolution_kernel(radius int) [][]float64 {
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

	return conv_kernel
}

func gaussian_blur_worker(jobChan chan matrix_elem, resultChan chan Pixel, loadedImage image.Image, conv_kernel [][]float64, radius, width, height int) {

	for running {
		elem := <-jobChan
		i := elem.index_i
		j := elem.index_j

		var red_output float64
		var green_output float64
		var blue_output float64
		var a_output float64

		for x := -radius; x < radius; x++ {
			for y := -radius; y < radius; y++ {
				gaussian_coefficient := conv_kernel[x+radius][y+radius]
				x_substitue := x
				y_substitue := y
				if i+x_substitue < 0 {
					x_substitue = -i
				}
				if j+y_substitue < 0 {
					y_substitue = -j
				}
				if i+x_substitue >= width {
					x_substitue = width - i - 1
				}
				if j+y_substitue >= height {
					y_substitue = height - j - 1
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
		color := color.NRGBA{
			R: uint8(red_output),
			G: uint8(green_output),
			B: uint8(blue_output),
			A: uint8(a_output),
		}

		pixel := Pixel{
			index_i: elem.index_i,
			index_j: elem.index_j,
			pigment: color,
		}

		resultChan <- pixel

	}

}

func main() {
	var loadedImage image.Image
	arguments := os.Args[1:]
	running = true
	numGoroutines = 30
	fileName := arguments[0]
	// Read image from file that already exists
	existingImageFile, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	fileExtension := filepath.Ext(fileName)
	fmt.Println(fileName, fileExtension)

	defer existingImageFile.Close()
	if fileExtension == ".png" {
		loadedImage, err = png.Decode(existingImageFile)
		if err != nil {
			log.Fatal(err)
		}
	} else if fileExtension == ".jpg" {
		loadedImage, err = jpeg.Decode(existingImageFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	width := loadedImage.Bounds().Max.X
	height := loadedImage.Bounds().Max.Y
	numJobs := 20
	jobChan := make(chan matrix_elem, numJobs)
	resultChan := make(chan Pixel, numJobs)

	output_img := image.NewNRGBA(image.Rect(0, 0, width, height))

	if len(arguments) == 0 {
		fmt.Println("Provide at least 2 arguments: the image path and the effect desired")
		os.Exit(1)
	}

	if arguments[1] == "grayscale" {

		for x := 0; x <= numGoroutines; x++ {
			go grayscale_worker(jobChan, resultChan, loadedImage)
		}

		go func() {
			for i := 0; i <= width; i++ {
				for j := 0; j <= height; j++ {
					elem := matrix_elem{
						index_i: i,
						index_j: j,
					}
					jobChan <- elem
				}
			}
		}()

		for i := 0; i <= width; i++ {
			for j := 0; j <= height; j++ {
				r := <-resultChan
				output_img.Set(r.index_i, r.index_j, r.pigment)
			}
		}

		running = false
		close(jobChan)

	} else if arguments[1] == "gaussian_blur" {

		if len(arguments) < 3 {
			fmt.Println("Provide a valid radius for the gaussian blur")
			fmt.Println("The radius determines the strength of the blur. The bigger the radius, the stronger the blur")
			os.Exit(1)
		}
		gaussian_radius, err := strconv.Atoi(arguments[2])
		if err != nil {
			fmt.Println("Provide a valid radius for the gaussian blur")
			fmt.Println("The radius determines the strength of the blur. The bigger the radius, the stronger the blur")
			os.Exit(1)
		}

		conv_kernel := create_convolution_kernel(gaussian_radius)

		for x := 0; x <= numGoroutines; x++ {
			go gaussian_blur_worker(jobChan, resultChan, loadedImage, conv_kernel, gaussian_radius, width, height)
		}

		go func() {
			for i := 0; i <= width; i++ {
				for j := 0; j <= height; j++ {
					elem := matrix_elem{
						index_i: i,
						index_j: j,
					}
					jobChan <- elem
				}
			}
		}()

		for i := 0; i <= width; i++ {
			for j := 0; j <= height; j++ {
				r := <-resultChan
				output_img.Set(r.index_i, r.index_j, r.pigment)
			}
		}

		running = false
		close(jobChan)
		// close(resultChan)

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
