package gork

import (
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash/adler32"
	"net"
	"net/netip"

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

func (list PeerList) ToMap() peerListMap {
	mp := make(peerListMap, len(list))
	for _, p := range list {
		props := make([][2]string, 0, p.Properties.Length())
		for k, v := range p.Properties.Entries() {
			props = append(props, [2]string{k, v})
		}
		mp[p.ToHex()] = props
	}
	return mp
}

func (m peerListMap) ToList() *PeerList {
	list := make(PeerList, 0, len(m))
	for hexkey, props := range m {
		p := NewPeer(delphi.KeyFromHex(hexkey).Bytes())
		for _, prop := range props {
			p.Properties.Set(prop[0], prop[1])
		}
		list = append(list, p)
	}
	return &list
}

// peerlist as it appears in a config
type peerListMap map[string][][2]string

func (pl PeerList) MarshalJSON() ([]byte, error) {
	m := make(peerListMap, len(pl))
	for _, peer := range pl {
		peer.Expand()
		m[peer.Key.ToHex()] = [][2]string{}
		for k, v := range peer.Properties.Entries() {
			m[peer.Key.ToHex()] = append(m[peer.Key.ToHex()], [2]string{k, v})
		}
	}
	return json.Marshal(m)
}

func (peerListPointer *PeerList) UnmarshalJSON(b []byte) error {
	pl := *peerListPointer
	pm := peerListMap{}
	err := json.Unmarshal(b, &pm)
	if err != nil {
		return err
	}
	for key, props := range pm {
		p := NewPeer(delphi.KeyFromHex(key).Bytes())
		for _, prop := range props {
			p.Properties.Set(prop[0], prop[1])
		}
		pl = append(pl, p)
	}
	return nil
}

func (p Peer) Address() (*net.UDPAddr, error) {

	addrStr, exists := p.Properties.Get("addr")
	if !exists {
		return nil, errors.New("address not found")
	}

	ap, err := netip.ParseAddrPort(addrStr)
	if err != nil {
		return nil, err
	}
	sendAddr := net.UDPAddrFromAddrPort(ap)
	return sendAddr, nil
}

// Expand sets inferred properties
func (p Peer) Expand() {
	p.Properties.Set("nick", p.Nickname())
	p.Properties.Set("grip", p.Grip())
	// p.Properties.MoveToFront("nick")
	// p.Properties.MoveToFront("grip")
}

// Contract deletes inferred keys
func (p Peer) Contract() {
	p.Properties.Delete("nick")
	p.Properties.Delete("grip")
}

// func asMap(kv *KV) map[string]string {
// 	m := make(map[string]string, kv.Len())
// 	for pair := kv.Oldest(); pair != nil; pair = pair.Next() {
// 		m[pair.Key] = pair.Value
// 	}
// 	return m
// }

func (p Peer) Config() (string, map[string]string) {
	k := p.Key.ToHex()
	p.Expand()
	m := p.Properties.AsMap()
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
	headers := p.Properties.AsMap()

	headers["grip"] = p.Grip()
	block := &pem.Block{
		Type:    "GORACLE PUBLIC KEY",
		Headers: headers,
		Bytes:   p.Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}
