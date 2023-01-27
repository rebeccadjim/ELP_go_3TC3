package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"sync"
)

var running bool

type matrix_elem struct {
	index_i int
	index_j int
}

type Pixel struct {
	index_i int
	index_j int
	pigment color.NRGBA
}

func grayscale_worker(jobChan chan matrix_elem, resultChan chan Pixel, loadedImage image.Image, wg *sync.WaitGroup) {
	//loops through the whole array (all the pixels of the image) and calculates the average value of r, g, b
	//by setting the average value as the value for each color we have a grayscale image
	wg.Add(1)
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
	wg.Done()
}

func main() {
	running = true
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

	numJobs := 20
	jobChan := make(chan matrix_elem, numJobs)
	resultChan := make(chan Pixel, numJobs)
	var wg sync.WaitGroup

	output_img := image.NewNRGBA(image.Rect(0, 0, width, height))

	arguments := os.Args[1:]
	if len(arguments) == 0 {
		fmt.Println("Provide at least on argument: the effect desired")
		os.Exit(1)
	}

	if arguments[0] == "grayscale" {

		numGoroutines := 10

		for x := 0; x <= numGoroutines; x++ {
			go grayscale_worker(jobChan, resultChan, loadedImage, &wg)
		}

		//
		go func() {
			for i := 0; i <= height; i++ {
				for j := 0; j <= width; j++ {
					elem := matrix_elem{
						index_i: i,
						index_j: j,
					}
					jobChan <- elem
				}
			}
		}()
		for i := 0; i <= height; i++ {
			for j := 0; j <= width; j++ {
				r := <-resultChan
				output_img.Set(r.index_i, r.index_j, r.pigment)
			}
		}
		running = false
		// wg.Wait()
		close(jobChan)

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
