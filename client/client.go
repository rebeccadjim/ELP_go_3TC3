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
	"path/filepath"
)

func map_int(to_map int, max_from int, max_to int) int {
	return (to_map * max_to) / max_from
}

func TCP_client(input_image image.Image, width, height int) *image.NRGBA {
	servAddr := "localhost:8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	send_image(conn, input_image, width, height)

	defer conn.Close()

	return receive_image(conn, width, height)
}


func send_image(conn net.Conn, input_image image.Image, width, height int) {
	var msg string
	buffer := make([]byte, 2048)
	msg = "dimensions " + strconv.Itoa(width) + " " + strconv.Itoa(height) + "\n"
	_, err := conn.Write([]byte(msg))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	count := 0
	msg = ""

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {

			imageColors := input_image.At(i, j)
			R_i, G_i, B_i, A_i := imageColors.RGBA()

			R := uint32(map_int(int(R_i), 65535, 255))
			G := uint32(map_int(int(G_i), 65535, 255))
			B := uint32(map_int(int(B_i), 65535, 255))
			A := uint32(map_int(int(A_i), 65535, 255))

			msg += strconv.Itoa(i) + " " + strconv.Itoa(j) + " "
			msg += strconv.FormatUint(uint64(R), 10) + " "
			msg += strconv.FormatUint(uint64(G), 10) + " "
			msg += strconv.FormatUint(uint64(B), 10) + " "
			msg += strconv.FormatUint(uint64(A), 10) + "\n"

			count++

			if count > 64 || (i == width-1 && j == height-1) {
				_, err = conn.Write([]byte(msg))
				if err != nil {
					println("Write to server failed:", err.Error())
					os.Exit(1)
				}
				count = 0
				msg = ""
				conn.Read(buffer)
			}
		}

	}
	msg = "end\n"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}
}

func receive_image(conn net.Conn, width, height int) *image.NRGBA {
	opened := true
	var data_array, msg_array []string
	output_img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for opened {
		var buffer2 = make([]byte, 2048)
		_, err := conn.Read(buffer2)

		if err != nil {
			fmt.Println("error read client")
			opened = false
		}

		msg := string(buffer2)

		data_array = strings.Split(msg, "\n")

		for k := 0; k < len(data_array); k++ {
			s := data_array[k]
			msg_array = strings.Split(s, " ")

			if msg_array[0] == "end" {
				opened = false
			} else if msg_array[0] == "dimensions" {
				continue
			} else if len(msg_array) == 6 {
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

				output_img.Set(i, j, color.NRGBA{
					R: uint8(R),
					G: uint8(G),
					B: uint8(B),
					A: uint8(A),
				})
			}
		}
		conn.Write([]byte("ok"))
	}
	return output_img
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Provide the image path as an argument")
	}

	// Read image from file that already exists
	existingImageFile, err := os.Open(os.Args[1])
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

	output_image := TCP_client(loadedImage, width, height)

	f, err := os.Create(existingImageFile.Name()[:len(existingImageFile.Name()) - len(filepath.Ext(existingImageFile.Name()))] + "_processed.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, output_image); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
