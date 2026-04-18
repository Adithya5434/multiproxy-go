package router

import (
	"fmt"
	"net"
	"io"
	"bytes"

	"multiproxy/utils"
	"multiproxy/protocols"
)

// var socks5Proxy = protocols.NewSocks5Proxy("", "")
// var socks4Proxy = protocols.NewSocks4Proxy("")
// var httpProxy = protocols.NewHTTPProxy("", "", "")
// var minecraftProxy = protocols.NewMCProxy("localhost", 25565)

func Route(
    conn net.Conn, socks5 *protocols.Proxy, socks4 *protocols.Proxy, http *protocols.Proxy, mc *protocols.MCProxy ) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("read error:", err)
		return
	}

	data := buf[:n]

	protocol := utils.DetectProtocol(data)
	fmt.Println("Detected protocol:", protocol)

	reader := io.MultiReader(bytes.NewReader(data), conn)


	switch protocol {
	case "socks5":
		socks5.HandleSocks5Client(reader, conn)
	
	case "socks4":
		socks4.HandleSocks4Client(reader, conn)

	case "http_proxy":
		http.HandleHTTPConnection(reader, conn)

	case "minecraft":
		mc.HandleMinecraftClient(reader, conn)
	}
}
