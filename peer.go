package gork

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash/adler32"

	"github.com/eloonstra/go-little-drunken-bishop/pkg/drunkenbishop"
	"github.com/goombaio/namegenerator"
	"github.com/sean9999/go-delphi"
	"github.com/vmihailenco/msgpack/v5"
)

type Serde interface {
	Serialize() []byte
	Deserialize(b []byte) error
}

type IPeer interface {
	Serde
	Id() []byte
	Key() delphi.Key
	Equal(IPeer) bool
}

type Peer struct {
	delphi.Key `msgpack:"pub" json:"pub" yaml:"pub"`
	Properties *KV `msgpack:"props" json:"props" yaml:"yaml"`
}

type IPeerList interface {
	Serde
	MarshalJSON() ([]byte, error)
	Get(k delphi.Key) (Peer, bool)
	Set(p Peer) bool
	Len() int
}

type PeerList []Peer

func (pl PeerList) MarshalJSON() ([]byte, error) {
	m := map[string]*KV{}
	for _, peer := range pl {
		peer.Expand()
		m[peer.ToHex()] = peer.Properties
	}
	return json.Marshal(m)
}

// Expand sets inferred properties
func (p Peer) Expand() {
	p.Properties.Set("nick", p.Nickname())
	p.Properties.Set("grip", p.Grip())
	p.Properties.MoveToFront("nick")
	p.Properties.MoveToFront("grip")
}

// Contract deletes inferred keys
func (p Peer) Contract() {
	p.Properties.Delete("nick")
	p.Properties.Delete("grip")
}

func asMap(kv *KV) map[string]string {
	m := make(map[string]string, kv.Len())
	for pair := kv.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m
}

func (p Peer) Config() (string, map[string]string) {
	k := p.Key.ToHex()
	p.Expand()
	m := asMap(p.Properties)
	return k, m
}

func NewPeer(b []byte) Peer {
	k := delphi.KeyFromBytes(b)
	props := NewKV()
	p := Peer{k, props}
	return p
}

// func (p *Peer) MarshalJSON() ([]byte, error) {
// 	m := p.Properties.AsMap()
// 	m["pub"] = p.Key.ToHex()
// 	m["grip"] = p.Grip()
// 	return json.Marshal(m)
// }

// func (p *Peer) UnmarshalJSON(b []byte) error {
// 	var m map[string]string
// 	err := json.Unmarshal(b, &m)
// 	if err != nil {
// 		return err
// 	}
// 	pubhex, exists := m["pub"]
// 	if !exists {
// 		return errors.New("no pub key")
// 	}
// 	pubkey := delphi.KeyFromHex(pubhex)
// 	delete(m, "pub")
// 	p.Key = pubkey
// 	p.Properties.Incorporate(m)
// 	return nil
// }

func (p Peer) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(p)
}

func (p Peer) UnmarshalBinary(b []byte) error {
	return msgpack.Unmarshal(b, p)
}

func (p Peer) Equal(q Peer) bool {
	return p.Key.Equal(q.Key)
}

func (p Peer) Nickname() string {
	seed := p.ToInt64()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	name := nameGenerator.Generate()
	return name
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
}

// MarshalPEM marshals a PEM to a Peer.
func (p Peer) MarshalPEM() ([]byte, error) {

	p.Expand()
	headers := asMap(p.Properties)

	headers["grip"] = p.Grip()
	block := &pem.Block{
		Type:    "GORACLE PUBLIC KEY",
		Headers: headers,
		Bytes:   p.Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}
