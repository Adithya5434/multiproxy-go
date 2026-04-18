package protocols

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"multiproxy/utils"
)



func NewMCProxy(host string, port int) *MCProxy {
	return &MCProxy{
		MCHost: host,
		MCPort: port,
	}
}

func (p *MCProxy) HandleMinecraftClient(reader io.Reader, conn net.Conn) {
	defer conn.Close()

	// Read initial packet
	var rawPacket = []byte{}
	tmp := make([]byte, 1024)

	var headerLen int
	var packetLength int
	var err error

	// read full packet
	for {
		n, err := reader.Read(tmp)
		if err != nil {
			fmt.Println("[ERROR][mc] Error reading client packet:", err)
			return
		}

		rawPacket = append(rawPacket, tmp[:n]...)

		packetLength, headerLen, err = utils.ReadVarInt(rawPacket, 0)
		// fmt.Println("[DEBUG][mc] packetLength:", packetLength, "headerLen:", headerLen)
		if err != nil {
			continue
		}

		if len(rawPacket) >= headerLen+packetLength {
			break
		}
	}

	// BYPASS TCPshield plugin
    // how TCPshield plugin works:
    // It checks its server ip in initial handshake (0x00 packet) from client
    // and if it doesn't match with the real server ip, it closes the connection.
    // So we need to send the real server ip and port in the initial handshake packet to bypass it.

	// parse initial handshake packet (0x00)
	i := headerLen

	packetID, ni, err := utils.ReadVarInt(rawPacket, i)
	if err != nil {
		fmt.Println("[ERROR][mc] Error reading packet ID:", err)
		return
	}
	i = ni

	if packetID != 0x00 {
		fmt.Printf("[ERROR][mc] Unexpected packet ID %d, expected 0x00 handshake\n", packetID)
		return
	}

	protocolVersion, ni, err := utils.ReadVarInt(rawPacket, i)
	if err != nil {
		fmt.Println("[ERROR][mc] Error reading protocol version:", err)
		return
	}
	i = ni

	addrLen, ni, err := utils.ReadVarInt(rawPacket, i)
	if err != nil {
		return
	}
	i = ni

	if i+addrLen > len(rawPacket) {
		fmt.Println("[ERROR][mc] Invalid address length")
		return
	}

	serverAddress := string(rawPacket[i : i+addrLen])
	i += addrLen

	if i+2 > len(rawPacket) {
		fmt.Println("[ERROR][mc] Invalid port data")
		return
	}

	serverPort := int(rawPacket[i])<<8 | int(rawPacket[i+1])
	i += 2

	intent, ni, err := utils.ReadVarInt(rawPacket, i)
	if err != nil {
		return
	}
	i = ni

	fmt.Println("[DEBUG][mc] Parsed:", protocolVersion, serverAddress, serverPort, intent)


	//  construct new handshake packet (0x00) with real host and port
	payload := []byte{}

	payload = append(payload, utils.WriteVarInt(0x00)...)
	payload = append(payload, utils.WriteVarInt(protocolVersion)...)
	payload = append(payload, utils.WriteVarInt(len(p.MCHost))...)
	payload = append(payload, []byte(p.MCHost)...)
	payload = append(payload, byte(p.MCPort>>8), byte(p.MCPort)) // big endian
	payload = append(payload, utils.WriteVarInt(intent)...)

	packet := append(utils.WriteVarInt(len(payload)), payload...)

	addr := net.JoinHostPort(p.MCHost, strconv.Itoa(p.MCPort))

	remote, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("[ERROR][mc] Dial error:", err)
		return
	}

	_, err = remote.Write(packet)
	if err != nil {
		fmt.Println("[ERROR][mc] Error sending handshake:", err)
		return
	}

	leftover := rawPacket[headerLen+packetLength:]
	if len(leftover) > 0 {
		_, err = remote.Write(leftover)
		if err != nil {
			fmt.Println("[ERROR][mc] Error sending leftover:", err)
			return
		}
	}

	relay(conn, remote)

}