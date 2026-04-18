package utils

import (
	"bytes"
	"strings"
)



func DetectProtocol(data []byte) (string) {
	if len(data) < 2{
        return "unknown"  // Too small to determine
	}

	// Minecraft
	if isMinecraftProtocol(data) {
		return "minecraft"
	}

	// SOCKS5
	if data[0] == 0x05 {
		return "socks5"
	}

	// SOCKS4
	if data[0] == 0x04 {
		return "socks4"
	}

	// HTTP - proxy and Web
	proto := detectHTTP(data)
	if proto == "http_proxy" {
		return "http_proxy"
	} else if proto == "http_web" {
		return "http_web"
	}

	return "unknown"
}

func detectHTTP(data []byte) string {
	// HTTP proxy
	if len(data) >= 8 && string(data[:8]) == "CONNECT " {
		return "http_proxy"
	}

	httpMethods := [][]byte{
		[]byte("GET "),
		[]byte("POST "),
		[]byte("PUT "),
		[]byte("DELETE "),
		[]byte("HEAD "),
		[]byte("OPTIONS "),
		[]byte("TRACE "),
		[]byte("PATCH "),
	}

	// any(data.startswith(method) for method in http_methods)
	matched := false
	for _, method := range httpMethods {
		if len(data) >= len(method) && bytes.HasPrefix(data, method) {
			matched = true
			break
		}
	}

	if !matched {
		return ""
	}

	// data.decode(errors='ignore')
	line := string(data)

	// .split('\r\n')[0]
	if idx := strings.Index(line, "\r\n"); idx != -1 {
		line = line[:idx]
	}

	// parts = line.split(" ")
	parts := strings.Split(line, " ")

	if len(parts) >= 2 {
		path := parts[1]

		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			return "http_proxy"
		}

		return "http_web"
	}

	return ""
}



func isMinecraftProtocol(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Legacy ping / handshake magic bytes
	if data[0] == 0xFE || data[0] == 0x02 || data[0] == 0x10 {
		return true
	}

	// Read packet length
	_, index, err := ReadVarInt(data, 0)
	if err != nil {
		return false
	}

	// Read packet ID
	packetID, _, err := ReadVarInt(data, index)
	if err != nil {
		return false
	}

	return packetID == 0x00
}
