package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

    conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
    buffer := make([]byte, 1024)
    _, err = conn.Read(buffer)
    if err != nil {
        fmt.Println("Error reading request")
    }
    reqLines := strings.Split(string(buffer), "\r\n")
    reqLine := strings.Split(reqLines[0], " ")
    urlPath := reqLine[1];

    var response string
    if (string(urlPath) == "/") {
        response = "HTTP/1.1 200 OK\r\n\r\n"
    } else if (strings.HasPrefix(urlPath, "/echo/")) {
        echo := urlPath[6:]
        response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
    } else {
        response = "HTTP/1.1 404 Not Found\r\n\r\n"
    }
    _, err = conn.Write([]byte(response))
    if err != nil {
        fmt.Println("Error writing data: ", err.Error())
        os.Exit(1)
    }
}
