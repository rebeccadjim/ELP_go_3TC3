package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
)

var first_output_img *image.NRGBA
var final_output_img *image.NRGBA

var numGoroutines = 8

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

func map_int(to_map int, max_from int, max_to int) int {
	return (to_map * max_to) / max_from
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

func handle_connection(conn net.Conn) {
	var data_array, msg_array []string
	opened := true
	var width, height int

	for opened {
		rec := make([]byte, 1024)

		_, err := conn.Read(rec)

		if err != nil {
			fmt.Println("error read server")
			opened = false
		}
		msg := string(rec)

		data_array = strings.Split(msg, "\n")

		for i := 0; i < len(data_array); i++ {
			s := data_array[i]
			msg_array = strings.Split(s, " ")

			if msg_array[0] == "end" {
				fmt.Println("Reached end")
				opened = false
			} else if msg_array[0] == "dimensions" {
				width, err = strconv.Atoi(msg_array[1])
				if err != nil {
					fmt.Println("error converting width")
				}
				height, err = strconv.Atoi(msg_array[2])
				if err != nil {
					fmt.Println("error converting height")
				}
				first_output_img = image.NewNRGBA(image.Rect(0, 0, width, height))

			} else if len(msg_array) > 1 {
				i, err := strconv.Atoi(msg_array[0])
				if err != nil {
					fmt.Println("error converting i")
				}
				j, err := strconv.Atoi(msg_array[1])
				if err != nil {
					fmt.Println("error converting j")
				}
				R, err := strconv.ParseUint(msg_array[2], 10, 64)
				if err != nil {
					fmt.Println("error converting R")
				}
				G, err := strconv.ParseUint(msg_array[3], 10, 64)
				if err != nil {
					fmt.Println("error converting G")
				}
				B, err := strconv.ParseUint(msg_array[4], 10, 64)
				if err != nil {
					fmt.Println("error converting B")
				}
				A, err := strconv.ParseUint(msg_array[5], 10, 64)
				if err != nil {
					fmt.Println("error converting A")
				}

				first_output_img.Set(i, j, color.NRGBA{
					R: uint8(R),
					G: uint8(G),
					B: uint8(B),
					A: uint8(A),
				})
				conn.Write([]byte("ok"))
			}
		}
	}

	fmt.Println("Reached end of read")
	f, err := os.Create("image.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, first_output_img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	// var loadedImage image.Image

	// existingImageFile, err := os.Open("image.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // loadedImage, err = png.Decode(existingImageFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	numJobs := 20
	jobChan := make(chan matrix_elem, numJobs)
	resultChan := make(chan Pixel, numJobs)

	final_output_img = image.NewNRGBA(image.Rect(0, 0, width, height))

	gaussian_radius := 10

	conv_kernel := create_convolution_kernel(gaussian_radius)
	running = true
	for x := 0; x <= numGoroutines; x++ {
		go gaussian_blur_worker(jobChan, resultChan, first_output_img, conv_kernel, gaussian_radius, width, height)
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
			final_output_img.Set(r.index_i, r.index_j, r.pigment)
		}
	}

	running = false
	close(jobChan)
	fmt.Println("Reached end of go routines")
	f_final, err := os.Create("image_processed.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f_final, final_output_img); err != nil {
		f_final.Close()
		log.Fatal(err)
	}

	if err := f_final.Close(); err != nil {
		log.Fatal(err)
	}

	var msg string
	buffer := make([]byte, 1024)
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			imageColors := final_output_img.At(i, j)
			R_i, G_i, B_i, A_i := imageColors.RGBA()

			R := uint32(map_int(int(R_i), 65535, 255))
			G := uint32(map_int(int(G_i), 65535, 255))
			B := uint32(map_int(int(B_i), 65535, 255))
			A := uint32(map_int(int(A_i), 65535, 255))

			msg = strconv.Itoa(i) + " " + strconv.Itoa(j) + " "
			msg += strconv.FormatUint(uint64(R), 10) + " "
			msg += strconv.FormatUint(uint64(G), 10) + " "
			msg += strconv.FormatUint(uint64(B), 10) + " "
			msg += strconv.FormatUint(uint64(A), 10) + "\n"

			_, err = conn.Write([]byte(msg))
			if err != nil {
				println("Write to server failed:", err.Error())
				os.Exit(1)
			}

			fmt.Println(msg)

			conn.Read(buffer)
			// if msg_rec != "ok" {
			// 	break
			// }
			// fmt.Println(msg)
		}
	}
	msg = "end\n"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}
}

func TCP_server() *image.NRGBA {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("error listen")
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accept")
		}
		go handle_connection(conn)
	}
}

func main() {
	TCP_server()
}
