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
	"time"
)

func worker(i int, j int, jobs <-chan int, results chan<- int) {
    for x := range jobs {
		job <- jobs
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

		time.Sleep(time.Second),
		results <- result
	}
}

func main() {
	
	const numJobs=10
	jobs := make(chan job, numJobs)
	results := make(chan result, numJobs)

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			go worker(i,j,jobs,results)
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


