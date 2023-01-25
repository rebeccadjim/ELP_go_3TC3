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

var output_img *image.NRGBA

func handle_connection(conn net.Conn) {
	var data_array, msg_array []string
	opened := true
	var width, height int

	for opened {
		rec := make([]byte, 1024)

		_, err := conn.Read(rec)

		if err != nil {
			fmt.Println("error read server")
		}
		msg := string(rec)
		data_array = strings.Split(msg, "\n")
		fmt.Println("dataray", data_array[:])

		for i := 0; i < len(data_array); i++ {
			s := data_array[i]
			msg_array = strings.Split(s, " ")

			if msg_array[0] == "end" {
				opened = false
				break
			} else if msg_array[0] == "dimensions" {
				width, err = strconv.Atoi(msg_array[1])
				if err != nil {
					fmt.Println("error converting width")
				}
				height, err = strconv.Atoi(msg_array[2])
				if err != nil {
					fmt.Println("error converting height")
				}
				output_img = image.NewNRGBA(image.Rect(0, 0, width, height))
				fmt.Println("Dimensions ok:", width, height)

			} else if len(msg_array) > 1 {
				fmt.Println("inelse", msg_array[:], len(msg_array))
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
