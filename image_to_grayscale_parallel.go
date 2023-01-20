/*
That's basically just the first idea, I still have to understand some of the code later because I haven't gone through it all properly yet
also see about other file formats...
and maybe other transformation
*/

package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func worker_grayscale(jobs chan *image.Image, loadedImage image.Image, output_img *image.NRGBA, width int, counter int) {
	//loops through the whole array (all the pixels of the image) and calculates the average value of r, g, b
	//by setting the average value as the value for each color we have a grayscale image
	for i := 0; i < width; i++ {
		colors := loadedImage.At(i, width)
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
	jobs <- output_img
}

func main() {

	const numJobs = 10
	jobs := make(chan job, numJobs)
	results := make(chan result, numJobs)

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			go worker(i, j, jobs, results)
		}
	}

	existingImageFile, err := os.Open("pic.png")
	if err != nil {
		log.Fatal(err)
	}

	defer existingImageFile.Close()

	loadedImage, err := png.Decode(existingImageFile)
	if err != nil {
		log.Fatal(err)
	}

	colors := loadedImage.At(loadedImage.Bounds().Max.X-1, loadedImage.Bounds().Max.Y-1)
	width := loadedImage.Bounds().Max.X
	height := loadedImage.Bounds().Max.x
	output_img := image.NewNRGBA(image.Rect(0, 0, width, height))

	f, err := os.Create("image.png")
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
}

func map_int(to_map int, max_from int, max_to int) int {
	return (to_map * max_to) / max_from
}
