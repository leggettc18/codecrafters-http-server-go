package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
    headerMap map[string]string
    url string
}

func Parse(buffer []byte) (Request, error) {
    reqLines := strings.Split(string(buffer), "\r\n")
    headers := reqLines[1:len(reqLines)-2]
    headerMap := make(map[string]string)
    for _, header := range headers {
        values := strings.Split(header, ": ")
        headerMap[values[0]] = values[1]
    }
    reqLine := strings.Split(reqLines[0], " ")
    urlPath := reqLine[1];
    return Request{ headerMap: headerMap, url: urlPath }, nil
}

func main() {
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

    request, _ := Parse(buffer)

    var response string
    if (request.url == "/") {
        response = "HTTP/1.1 200 OK\r\n\r\n"
    } else if (strings.HasPrefix(request.url, "/echo/")) {
        echo := strings.TrimPrefix(request.url, "/echo/")
        response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
    } else if (request.url == "/user-agent") {
        userAgent := request.headerMap["User-Agent"]
        if (userAgent == "") {
            response = "HTTP/1.1 400 Bad Request\r\n\r\n"
        } else {
            response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
        }
    } else {
        response = "HTTP/1.1 404 Not Found\r\n\r\n"
    }
    _, err = conn.Write([]byte(response))
    if err != nil {
        fmt.Println("Error writing data: ", err.Error())
        os.Exit(1)
    }
}
