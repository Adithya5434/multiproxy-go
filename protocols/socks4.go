package protocols

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)


func NewSocks4Proxy(username string) *Proxy {
	return &Proxy{
		Username: username,
		Protected: username != "",
	}
}


func (p *Proxy) HandleSocks4Client(reader io.Reader, conn net.Conn) {
	fmt.Println("[DEBUG][s4] New SOCKS4 connection from:", conn.RemoteAddr())

	defer func() {
		fmt.Println("[DEBUG][s4] Closing SOCKS4 connection:", conn.RemoteAddr())
		conn.Close()
	}()

	// --- Handshake ---
	buf := make([]byte, 2)
	_, err := io.ReadFull(reader, buf) // read version + cmd
	if err != nil {
		fmt.Println("[ERROR][s4] Read handshake failed:", err)
		return
	}

	version := buf[0]
	command := buf[1]

	if version != 4 || command != 1 { // ver + cmd
		fmt.Println("[ERROR][s4] Invalid SOCKS4 request | version:", version, "cmd:", command)
		conn.Close()
		return
	}

	// port(2 bytes) + address(4 bytes)
	adrPortBuf := make([]byte, 6)
	_, err = io.ReadFull(reader, adrPortBuf) 
	if err != nil {
		fmt.Println("[ERROR][s4] Failed reading IP/port:", err)
		return
	}
	
	port := int(binary.BigEndian.Uint16(adrPortBuf[:2]))
	address := net.IP(adrPortBuf[2:6]).String()

	// read username
	userid, err := readUntilNull(reader)
	if err != nil {
		fmt.Println("[ERROR][s4] Failed reading USERID:", err)
		return
	}

	// authentication
	if p.Protected {
		if string(userid) != p.Username {
			fmt.Println("[WARN][s4] Authentication failed | provided:", string(userid))
			conn.Close()
			return
		}
	}

	// check if its socks4a
	if address[:6] == "0.0.0." {
		domain, err := readUntilNull(reader)

		if err != nil {
			fmt.Println("[ERROR][s4] Failed reading domain:", err)
			return
		}

		address = string(domain)

	}

	p.handlesocks4Request(conn, address, port)
}


func (p *Proxy) handlesocks4Request(conn net.Conn, address string, port int) {

	remote, err := net.Dial("tcp", net.JoinHostPort(address, fmt.Sprintf("%d", port)))
	if err != nil {
		fmt.Println("[ERROR][s4] Remote connection failed:", err)
		conn.Write([]byte{0x00, 0x5B, 0, 0, 0, 0, 0, 0})
		return
	}
	defer remote.Close()


	resp := make([]byte, 8)
	resp[0] = 0x00
	resp[1] = 0x5A // request granted
	binary.BigEndian.PutUint16(resp[2:4], uint16(port))
	copy(resp[4:8], net.ParseIP(address).To4())

	_, err = conn.Write(resp)
	if err != nil {
		fmt.Println("[ERROR][s4] Failed to send success response:", err)
		return
	}

	fmt.Println("[DEBUG][s4] SOCKS4 handshake complete, starting relay")
	relay(conn, remote)
}


func readUntilNull(r io.Reader) ([]byte, error) {
	var out []byte
	buf := []byte{0}

	for {
		n, err := r.Read(buf)
		if n == 1 {
			if buf[0] == 0x00 {
				return out, nil
			}
			out = append(out, buf[0])
		}
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
	}
}
