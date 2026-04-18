package protocols

import (
	"net"
	"io"
)


func relay(a, b net.Conn) {
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(a, b)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(b, a)
		done <- struct{}{}
	}()

	<-done
}