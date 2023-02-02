# Gaussian blur server

A server that takes an image from a client via a socket, and sends back the image blurred by a parallelized gaussian blur algorithm.

## Usage

### Server

To run the server:

```shell
cd server
go run server.go
``` 

### Client

To run the client:

```shell
cd client
go run client.go {image_path}
```

Where {image_path} is the path of the image you want to blur. It will then be saved in the current directory.

## Team
[Rebecca Djimtoingar](https://github.com/rebeccadjim) and [Valentin Jossic](https://github.com/vqlion)