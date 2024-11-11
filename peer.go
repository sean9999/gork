package gork

import (
	"encoding/pem"
	"fmt"
	"hash/adler32"

	"github.com/eloonstra/go-little-drunken-bishop/pkg/drunkenbishop"
	"github.com/sean9999/go-delphi"
	stablemap "github.com/sean9999/go-stable-map"
	"github.com/vmihailenco/msgpack/v5"
)

type Peer struct {
	delphi.Key `msgpack:"pub" json:"pub"`
	Properties *stablemap.StableMap[string, string] `msgpack:"props" json:"props"`
}

func (p Peer) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(p)
}

func (p Peer) UnmarshalBinary(b []byte) error {
	return msgpack.Unmarshal(b, p)
}

func (p Peer) Equal(q Peer) bool {
	return p.Key.Equal(q.Key)
}

// a key grip is a string short enough to be recognizable by the human eye
// and long enough to be reasonably unique
func (p Peer) Grip() string {
	s := adler32.Checksum(p.Bytes())
	//s := crc32.Checksum(p.Bytes(), crc32.IEEETable)
	return fmt.Sprintf("%x", s)
}

// Art returns ASCII art for a Peer
func (p Peer) Art() string {
	title := fmt.Sprintf("ORACLE PEER %s", p.Grip())
	return drunkenbishop.GenerateRandomArt(34, 18, p.Bytes(), true, title)
	//return randomart.RandomArt(p.Bytes(), p.Grip())
	//return drunkenbishop.GenerateHeatmap(36, 14, p.Bytes())
}

func (p Peer) MarshalPEM() ([]byte, error) {

	headers := p.Properties.AsMap()
	headers["grip"] = p.Grip()

	block := &pem.Block{
		Type:    "GORACLE PUBLIC KEY",
		Headers: headers,
		Bytes:   p.Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}
