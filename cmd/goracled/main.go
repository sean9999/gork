package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/spf13/afero"
)

type state struct {
	conf        gork.ConfigProvider
	port        uint
	self        *gork.Principal
	localAddr   net.Addr
	environment hermeti.Env
}

func main() {

	env := hermeti.RealEnv()
	env.Args = env.Args[1:]
	filesystem := afero.NewOsFs()
	exe, err := initialize(filesystem, env)
	if err != nil {
		log.Fatal(err)
	}

	me := exe.self

	fmt.Fprintln(env.OutStream, me.Nickname())
	io.Copy(env.OutStream, me.Export())

	// listen to incoming UDP packets
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", exe.port))
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	exe.localAddr = pc.LocalAddr()
	spool := NewSpool(pc)

	for {
		select {
		case inEnv := <-spool.inbox:
			//	do something with a well-formed message
			go processEnvelope(exe, inEnv, spool.errors, spool.outbox)

		case err := <-spool.errors:
			fmt.Println("error", err)
		case outEnv := <-spool.outbox:
			//spool.Send(outMsg, outMsg.ToPEM())
			fmt.Println(outEnv)
		}
	}

}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	pc.WriteTo(buf, addr)
}

// process an envelope and push messages to outbox and/or errs, if you want
func processEnvelope(s state, e Envelope, errs chan error, outbox chan Envelope) {

	switch e.Message.Subject {
	case "ASSERTION":
		processAssertion(s, e, errs, outbox)
	default:
		err := fmt.Errorf("unrecognized subject: %q", e.Message.Subject)
		errs <- err
	}

}
