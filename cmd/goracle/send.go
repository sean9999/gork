package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/hermeti"
)

func (exe *Exe) Send(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {
	args, err := exe.ensureSelf(ctx, env, args)
	if err != nil {
		return nil, fmt.Errorf("couldn't send: %w", err)
	}

	if len(args) < 1 {
		return nil, errors.New("you must pass address as arg")
	}
	ap, err := netip.ParseAddrPort(args[0])
	if err != nil {
		return nil, err
	}
	sendAddr := net.UDPAddrFromAddrPort(ap)
	args = args[1:]

	conn, err := net.DialUDP("udp", nil, sendAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	msg := delphi.NewMessage(env.Randomness, []byte("all your base are belong to us."))
	msg.Subject = "ALL BASE"

	msg.Sender = exe.Self.PublicKey()

	msg.Sign(env.Randomness, &exe.Self)

	msgAsBytes, err := io.ReadAll(msg)
	if err != nil {
		return nil, err
	}

	// msgBytes, err := msg.MarshalBinary()
	// if err != nil {
	// 	return nil, err
	// }

	// i, err := conn.WriteToUDP(msgAsBytes, sendAddr)

	i, err := conn.Write(msgAsBytes)

	if err != nil {
		return nil, err
	}

	fmt.Fprintf(env.OutStream, "%d bytes written\n", i)

	return args, nil

}
