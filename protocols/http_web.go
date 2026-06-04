package protocols

import (
	"fmt"
	"io"
	"net"
)

func NewHTTPWebProxy(host string, port int) *WebProxy {
	return &WebProxy{
		Host: host,
		Port: port,
	}
}

func (p *WebProxy) HandleHTTPWebConnection(reader io.Reader, conn net.Conn) {
    backend, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.Host, p.Port))
    if err != nil {
		fmt.Println("Failed to connect to backend HTTP Web server:", err)
        return
    }
    defer backend.Close()

	go io.Copy(backend, reader)
    io.Copy(conn, backend)
}