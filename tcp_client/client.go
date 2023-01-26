package main

import (
	// "fmt"
	// "fmt"
	"image"
	"image/png"
	"log"
	"net"
	"os"
	"strconv"
	// "reflect"
)

func TCP_client(image image.Image, width, height int) {
	servAddr := "localhost:8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	var msg string
	buffer := make([]byte, 1024)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	msg = "dimensions " + strconv.Itoa(width) + " " + strconv.Itoa(height) + "\n"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			imageColors := image.At(i, j)
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

			conn.Read(buffer)
			// if msg_rec != "ok" {
			// 	break
			// }
			// fmt.Println(msg)
		}
	}

	msg = "end"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	defer conn.Close()
}

func map_int(to_map int, max_from int, max_to int) int {
	return (to_map * max_to) / max_from
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

	TCP_client(loadedImage, width, height)

}
