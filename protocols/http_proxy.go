package protocols

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func NewHTTPProxy(host string, username string, password string) *Proxy {
	return &Proxy{
		Username: username, 							 // currently username and password
		Password: password, 							 // are not supported for HTTP proxy
		Protected: username != "" || password != "",     // will be added in the future
	}
}


func (p *Proxy) HandleHTTPConnection(reader io.Reader, conn net.Conn) {
	fmt.Println("[DEBUG][http] New connection from:", conn.RemoteAddr())
	defer func() {
		fmt.Println("[DEBUG][http] Closing connection:", conn.RemoteAddr())
		conn.Close()
	}()

	// --- Handshake ---
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil {
		fmt.Println("[ERROR][http] Read handshake failed:", err)
		return
	}

	data := buf[:n]
	text := string(data)
	
	firstLine := strings.SplitN(text, "\r\n", 2)[0]
	parts := strings.Split(firstLine, " ")

	if len(parts) < 2 {
		fmt.Println("[ERROR][http] Bad HTTP request (invalid first line): ", firstLine)
		return
	}

	method := strings.ToUpper(parts[0])
	target := parts[1]


	// HTTPs
	if method == "CONNECT" {
		remote, err := net.Dial("tcp", target)
		if err != nil {
			fmt.Println("[ERROR][http] Dial failed (CONNECT):", target, "| err:", err)
			return
		}

		_, err = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

		if err != nil {
			fmt.Println("[ERROR][http] Failed to send CONNECT response:", err)
			remote.Close()
			return
		}

		fmt.Println("[DEBUG][http] Tunnel established, starting relay")
		relay(conn, remote)


	// HTTP
	} else {
		lines := strings.Split(text, "\r\n")

		var hostLine string

		for _, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.HasPrefix(lineLower, "host:") {
				hostLine = line
				break
			}
		}

		if hostLine == "" {
			fmt.Println("[ERROR][http] No Host header found")
			return
		}

		host := strings.TrimSpace(hostLine[5:])
		port := "80" // set port 80 by default

		if strings.Contains(host, ":") {
			hp := strings.Split(host, ":")
			if len(hp) == 2 {
				host = hp[0]
				port = hp[1]
			}
		}

		remote, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			fmt.Println("[ERROR][http] Dial failed (HTTP):", host, port, "| err:", err)
			return
		}

		fmt.Println("[DEBUG][http] Connected to upstream:", remote.RemoteAddr())
		
		_, err = remote.Write(data)
		if err != nil {
			fmt.Println("[ERROR][http] Failed to forward request:", err)
			remote.Close()
			return
		}

		fmt.Println("[DEBUG][http] Request forwarded, starting relay")
		relay(conn, remote)
	}
}