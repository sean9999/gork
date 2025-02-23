package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/spf13/afero"
)

type state struct {
	conf        gork.ConfigProvider
	port        uint16
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

	//	if not explicitely set, try to get local address from config
	if addrStr, exists := exe.self.Props.Get("addr"); exists {
		if exe.port == 0 {

			//net.ResolveUDPAddr("udp6", addrStr)

			addr, err := netip.ParseAddrPort(addrStr)

			addr.Port()

			//addr, err := url.Parse(addrStr)
			if err == nil {

				exe.port = addr.Port()

				// p, err := strconv.Atoi(addr.Port())
				// if err == nil {
				// 	exe.port = uint(p)
				// }
			}
		}
	}

	// listen to incoming UDP packets
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", exe.port))
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	exe.localAddr = pc.LocalAddr()

	exe.self.Props.Set("addr", pc.LocalAddr().String())
	err = exe.self.Save(nil)
	if err != nil {
		log.Fatal(err)
	}

	spool := NewSpool(pc)

	for {
		select {
		case inEnv := <-spool.inbox:

			greenmsg := fmt.Sprintf("%s%s%s", Green, inEnv.Message.Subject, Reset)
			fmt.Fprintln(env.OutStream, greenmsg)
			//	do something with a well-formed message
			go processEnvelope(exe, inEnv, spool.errors, spool.outbox)

		case err := <-spool.errors:
			redmsg := fmt.Sprintf("%s%s%s", Red, err, Reset)
			fmt.Fprintln(env.ErrStream, redmsg)
		case outEnv := <-spool.outbox:
			//spool.Send(outMsg, outMsg.ToPEM())
			fmt.Println(outEnv)
		case logmsg := <-spool.logs:
			bluemsg := fmt.Sprintf("%s%s%s", Blue, logmsg, Reset)
			fmt.Fprintln(env.OutStream, bluemsg)
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
	case "ALL BASE":
		processAllBase(s, e, errs, outbox)
	default:
		err := fmt.Errorf("unrecognized subject: %q", e.Message.Subject)
		errs <- err
	}

}
