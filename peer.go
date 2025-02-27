package gork

import (
	"crypto/sha512"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash/adler32"
	"maps"

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
	Properties KV `msgpack:"props" json:"props" yaml:"yaml"`
}

func (p *Peer) MarshalJSON() ([]byte, error) {
	if p == nil {
		return nil, errors.New("nil peer")
	}
	obj := maps.Clone(p.Properties)
	obj["pubkey"] = p.ToHex()
	return json.MarshalIndent(obj, "", "\t")
}

func (p *Peer) UnmarshalJSON(b []byte) error {
	var obj map[string]string
	err := json.Unmarshal(b, obj)
	if err != nil {
		return err
	}
	for k, v := range obj {
		switch k {
		case "pubkey":
			p.Key = delphi.KeyFromHex(v)
		default:
			p.Properties[k] = v
		}
	}
	return nil
}

func (p *Peer) Digest() []byte {

	//	pubkey
	dig := make([]byte, 0)
	dig = append(dig, p.Key.Bytes()...)

	//	props without redundant derived keys
	props := p.Properties.WithoutInferred()
	for k, v := range props.LexicalOrder() {
		dig = append(dig, []byte(k)...)
		dig = append(dig, []byte(v)...)
	}

	//	hash it
	return sha512.New().Sum(dig)

}

// type IPeerList interface {
// 	Serde
// 	MarshalJSON() ([]byte, error)
// 	Get(k delphi.Key) (Peer, bool)
// 	Set(p Peer) bool
// 	Len() int
// }

// type PeerList []Peer

// func (list PeerList) ToMap() peerListMap {
// 	mp := make(peerListMap, len(list))
// 	for _, p := range list {
// 		props := make([][2]string, 0, p.Properties.Length())
// 		for k, v := range p.Properties.Entries() {
// 			props = append(props, [2]string{k, v})
// 		}
// 		mp[p.ToHex()] = props
// 	}
// 	return mp
// }

// func (m peerListMap) ToList() *PeerList {
// 	list := make(PeerList, 0, len(m))
// 	for hexkey, props := range m {
// 		p := NewPeer(delphi.KeyFromHex(hexkey).Bytes())
// 		for _, prop := range props {
// 			p.Properties.Set(prop[0], prop[1])
// 		}
// 		list = append(list, p)
// 	}
// 	return &list
// }

// // peerlist as it appears in a config
// type peerListMap map[string][][2]string

// func (pl PeerList) MarshalJSON() ([]byte, error) {
// 	m := make(peerListMap, len(pl))
// 	for _, peer := range pl {
// 		peer.Expand()
// 		m[peer.Key.ToHex()] = [][2]string{}
// 		for k, v := range peer.Properties.Entries() {
// 			m[peer.Key.ToHex()] = append(m[peer.Key.ToHex()], [2]string{k, v})
// 		}
// 	}
// 	return json.Marshal(m)
// }

// func (peerListPointer *PeerList) UnmarshalJSON(b []byte) error {
// 	pl := *peerListPointer
// 	pm := peerListMap{}
// 	err := json.Unmarshal(b, &pm)
// 	if err != nil {
// 		return err
// 	}
// 	for key, props := range pm {
// 		p := NewPeer(delphi.KeyFromHex(key).Bytes())
// 		for _, prop := range props {
// 			p.Properties.Set(prop[0], prop[1])
// 		}
// 		pl = append(pl, p)
// 	}
// 	return nil
// }

func NewPeer(b []byte) Peer {
	k := delphi.KeyFromBytes(b)
	props := make(KV)
	p := Peer{k, props}
	return p
}

// here we set our preferred binary serialization
func (p Peer) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(p)
}

// here we set our preferred binary de-serialization
func (p Peer) UnmarshalBinary(b []byte) error {
	return msgpack.Unmarshal(b, p)
}

// a Peer is equal to a Peer if its public key is the same
func (p Peer) Equal(q Peer) bool {
	return p.Key.Equal(q.Key)
}

// a Nickname is a very memorable string for humans only. Not to be used for actual uniqueness.
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
	return drunkenbishop.GenerateRandomArt(32, 16, p.Bytes(), true, title)
}

// MarshalPEM marshals a PEM to a Peer.
func (p Peer) MarshalPEM() ([]byte, error) {
	headers := p.Properties.WithInferred(p)
	block := &pem.Block{
		Type:    "GORACLE PEER",
		Headers: headers,
		Bytes:   p.Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}
