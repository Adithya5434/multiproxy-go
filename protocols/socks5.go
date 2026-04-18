package protocols

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)



func NewSocks5Proxy(username, password string) *Proxy {
	return &Proxy{
		Username: username,
		Password: password,
		Protected: username != "" || password != "",
	}
}

func (p *Proxy) HandleSocks5Client(reader io.Reader, conn net.Conn) {
	fmt.Println("[DEBUG][s5] New SOCKS5 connection from:", conn.RemoteAddr())

	defer func() {
		fmt.Println("[DEBUG][s5] Closing SOCKS5 connection:", conn.RemoteAddr())
		conn.Close()
	}()

	// initial handshake
	buf := make([]byte, 2)
	_, err := io.ReadFull(reader, buf) // read version + nMethods
	if err != nil {
		fmt.Println("[ERROR][s5] Handshake read failed:", err)
		return
	}
	
	ver := buf[0]
	nMethods := int(buf[1])
	
	fmt.Println("[DEBUG][s5] Version:", ver, "| Methods count:", nMethods)

	if ver != 5 {
		fmt.Println("[ERROR][s5] Invalid SOCKS version:", ver)
		return
	}

	methods := make([]byte, nMethods)
	_, err = io.ReadFull(reader, methods) // read supported methods
	if err != nil {
		fmt.Println("[ERROR][s5] Failed reading methods:", err)
		return
	}

	// authentication
	if p.Protected {
		ok := false
		for _, m := range methods {
			if m == 2 {
				ok = true
				break
			}
		}
		if !ok {
			conn.Write([]byte{5, 0xFF}) // NO ACCEPTABLE METHODS
			return
		}

		conn.Write([]byte{5, 2}) // USERNAME/PASSWORD AUTHENTICATION

		if !p.verifyCredentials(conn) {
			return
		}
	} else {
		conn.Write([]byte{5, 0}) // NO AUTHENTICATION REQUIRED
	}

	// --- Request ---
	p.handlesocks5Request(conn)
}

func (p *Proxy) HandleClient5(conn net.Conn) {
    defer conn.Close()

	buf := make([]byte, 2)
	n, err := io.ReadFull(conn, buf)
	fmt.Println(n)
	if err != nil {
		fmt.Println("[ERROR][s5] Failed reading handshake:", err)
		return
	}
	
	ver := buf[0]
	nMethods := int(buf[1])

	if ver != 5 {
		fmt.Println("[ERROR][s5] Invalid SOCKS version:", ver)
		return
	}

	methods := make([]byte, nMethods)
	_, err = conn.Read(methods)
	if err != nil {
		fmt.Println("[ERROR][s5] Failed reading methods:", err)
		return
	}

	if p.Protected {
		ok := false
		for _, m := range methods {
			if m == 2 {
				ok = true
				break
			}
		}

		if !ok {
			conn.Write([]byte{5, 0xFF}) // NO ACCEPTABLE METHODS
			return
		}

		conn.Write([]byte{5, 2}) // USERNAME/PASSWORD AUTHENTICATION

		if !p.verifyCredentials(conn) {
			return
		}

	} else {
		conn.Write([]byte{5, 0}) // NO AUTHENTICATION REQUIRED
	}
	
	p.handlesocks5Request(conn)

}



func (p *Proxy) handlesocks5Request(conn net.Conn) {
	header := make([]byte, 4)

	_, err := conn.Read(header)
	if err != nil {
		fmt.Println("[ERROR][s5] Failed reading request header:", err)
		return
	}

	ver, cmd, rsv, addrType  := header[0], header[1], header[2], header[3]
	
	if ver != 5 || rsv != 0 {
		fmt.Println("[ERROR][s5] Invalid request header")
		conn.Close()
	}

	if cmd != 1 { // 1-CONNECT 2-BIND  3-UDP ASSOCIATE
		fmt.Println("[WARN][s5] Unsupported command:", cmd)
		conn.Write([]byte{5, 7, 0, addrType, 0, 0, 0, 0 ,0 ,0}) // Command not supported
		conn.Close()
		return
	}

	if addrType != 1 && addrType != 3 {
		fmt.Println("[WARN][s5] Unsupported address type:", addrType)
		conn.Write([]byte{5, 8, 0, addrType, 0, 0, 0, 0 ,0 ,0}) // Address type not supported
		conn.Close()
		return
	}

	var addr string
	switch addrType {
		case 1: // IPv4
			ip := make([]byte, 4)
			conn.Read(ip)
			addr = net.IP(ip).String()
		
		case 3: // Domain
			lenBuf := make([]byte, 1)
			conn.Read(lenBuf)
			domainLen := int(lenBuf[0])
			domain := make([]byte, domainLen)
			conn.Read(domain)
			addr = string(domain)
		
		default:
			fmt.Println("Unsupported address type")
			return
	}

	portBuf := make([]byte, 2)
	conn.Read(portBuf)
	port := int(binary.BigEndian.Uint16(portBuf))

	remote, err := net.Dial("tcp", net.JoinHostPort(addr, fmt.Sprintf("%d", port)))
	if err != nil {
		fmt.Println("[ERROR][s5] Remote connection failed:", err)
		return
	}
	defer remote.Close()

	resp := []byte{5, 0, 0, 1}
	resp = append(resp, net.ParseIP("0.0.0.0").To4()...)
	resp = append(resp, 0, 0) // port 0
	conn.Write(resp) // success

	relay(conn, remote)
}


func (p *Proxy) verifyCredentials(conn net.Conn) bool {
	buff := make([]byte, 2)
	_, err := conn.Read(buff)
	if err != nil {
		return false
	}

	ulen := int(buff[1])
	uname := make([]byte, ulen)
	conn.Read(uname)

	plenBuff := make([]byte, 1)
	_, err = conn.Read(plenBuff)
	if err != nil {
		return false
	}

	plen := int(plenBuff[0])
	pname := make([]byte, plen)
	conn.Read(pname)

	if string(uname) == p.Username && string(pname) == p.Password {
		conn.Write([]byte{5, 0}) 
		fmt.Println("Client authenticated:", string(uname))
		return true
	}

	conn.Write([]byte{5, 0xFF})  // failed authentication
	return false
}


