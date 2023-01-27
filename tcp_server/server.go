package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var first_output_img *image.NRGBA
var final_output_img *image.NRGBA

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

func handle_connection(conn net.Conn) {
	var data_array, msg_array []string
	opened := true
	var width, height int

	numJobs := 20
	jobChan := make(chan matrix_elem, numJobs)
	resultChan := make(chan Pixel, numJobs)

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

			if msg_array[0] == "end\n" {
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

func TCP_server() *image.NRGBA {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("error listen")
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println("error accept")
		}
		go handle_connection(conn)
	}
}

func main() {
	TCP_server()
}
