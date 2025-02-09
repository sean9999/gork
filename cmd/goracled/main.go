package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/spf13/afero"
)

type state struct {
	conf fs.File
	port uint
	self *gork.Principal
}

func main() {

	env := hermeti.RealEnv()
	filesystem := afero.NewIOFS(afero.NewOsFs())
	exe, err := initialize(filesystem, os.Args[1:])

	if err != nil {
		panic(err)
	}

	fmt.Fprintln(env.OutStream, exe.self.AsPeer().Nickname())
	io.Copy(env.OutStream, exe.self.Export())

	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", exe.port))
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Println("error", err)
			continue
		}

		fmt.Fprintf(env.OutStream, "n = %d, addr = %s\n", n, addr)

		go serve(pc, addr, buf[:n])
	}

}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	// 0 - 1: ID
	// 2: QR(1): Opcode(4)
	buf[2] |= 0x80 // Set QR bit

	pc.WriteTo(buf, addr)
}
