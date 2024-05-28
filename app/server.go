package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strings"
)

type Request struct {
    method string
    headerMap map[string]string
    url string
    body string
}

type Response struct {
    code int
    message string
    headerMap map[string]string
    body string
}

func (self *Response) SetBody(body string) {
    if (body != "") {
        if (self.headerMap["Content-Encoding"] == "gzip") {
            var bodyBuf bytes.Buffer
            gzw := gzip.NewWriter(&bodyBuf)
            gzw.Write([]byte(body))
            gzw.Close()
            self.body = string(bodyBuf.Bytes())
        } else {
            self.body = body
        }
        self.headerMap["Content-Length"] = fmt.Sprintf("%d", len(self.body))
    }
}

func (self *Response) Print() string {
    response := fmt.Sprintf("HTTP/1.1 %d %s\r\n", self.code, self.message)
    for key, value := range self.headerMap {
        response += fmt.Sprintf("%s: %s\r\n", key, value)
    }
    response += "\r\n"
    if (self.body != "") {
        response += fmt.Sprintf("%s", self.body)
    }
    return response
}

func handleConnection(conn net.Conn) {
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Error reading request")
    }
    buffer = buffer[0:n]

    request, _ := Parse(buffer)

    response := Response{ code: 404, message: "Not Found", headerMap: make(map[string]string), body: "" }
    if (request.headerMap["Accept-Encoding"] != "") {
        encodings := strings.Split(request.headerMap["Accept-Encoding"], ", ")
        for _, encoding := range encodings {
            if (encoding == "gzip") {
                response.headerMap["Content-Encoding"] = encoding
            }
        }
    }
    if (request.method == "GET") {
        if (request.url == "/") {
            response.code = 200
            response.message = "OK"
        } else if (strings.HasPrefix(request.url, "/echo/")) {
            echo := strings.TrimPrefix(request.url, "/echo/")
            response.code = 200
            response.message = "OK"
            response.headerMap["Content-Type"] = "text/plain"
            response.SetBody(echo)
        } else if (strings.HasPrefix(request.url, "/files/")) {
            directory := os.Args[2]
            fileName := strings.TrimPrefix(request.url, "/files/")
            data, err := os.ReadFile(directory + fileName)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error reading file %s: %s", fileName, err.Error())
            } else {
                contents := string(data)
                response.code = 200
                response.message = "OK"
                response.headerMap["Content-Type"] = "application/octet-stream"
                response.SetBody(contents)
            }
        } else if (request.url == "/user-agent") {
            userAgent := request.headerMap["User-Agent"]
            if (userAgent == "") {
                response.code = 400
                response.message = "Bad Request"
            } else {
                response.code = 200
                response.message = "OK"
                response.headerMap["Content-Type"] = "text/plain"
                response.SetBody(userAgent)
            }
        }
    } else if (request.method == "POST") {
        if (strings.HasPrefix(request.url, "/files/")) {
            directory := os.Args[2]
            filename := strings.TrimPrefix(request.url, "/files/")
            os.WriteFile(directory + "/" + filename, []byte(request.body), fs.FileMode(777))
            response.code = 201
            response.message = "Created"

        }
    }
    _, err = conn.Write([]byte(response.Print()))
}

func Parse(buffer []byte) (Request, error) {
    rawBody := strings.Split(string(buffer), "\r\n\r\n")
    reqLines := strings.Split(rawBody[0], "\r\n")
    headers := reqLines[1:]
    headerMap := make(map[string]string)
    for _, header := range headers {
        values := strings.Split(header, ": ")
        headerMap[values[0]] = values[1]
    }
    reqLine := strings.Split(reqLines[0], " ")
    method := reqLine[0]
    urlPath := reqLine[1];
    body := ""
    if (len(rawBody) > 1) {
        body = rawBody[1];
    }
    return Request{ method: method, headerMap: headerMap, url: urlPath, body: body }, nil
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting connection: ", err.Error())
            os.Exit(1)
        }
        go handleConnection(conn)
    }
}
