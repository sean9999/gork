package gork

import (
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/sean9999/go-delphi"
)

type Peer struct {
	delphi.Key
}

func (p Peer) Equal(q Peer) bool {
	return p.Key.Equal(q.Key)
}

type Gork struct {
	delphi.Principal
	Peers []Peer
}

func NewGork(randy io.Reader) Gork {
	prince := delphi.NewPrincipal(randy)
	peers := make([]Peer, 0, 8)
	return Gork{*prince, peers}
}

func (g *Gork) HasPeer(p Peer) bool {
	return slices.Contains(g.Peers, p)
}

var ErrPeerExists = errors.New("peer already exists")

func (g *Gork) DropPeer(p Peer) {
	p2 := make([]Peer, 0, len(g.Peers))
	for _, thisPeer := range g.Peers {
		if !thisPeer.Equal(p) {
			p2 = append(p2, thisPeer)
		}
	}
	g.Peers = p2
}

func (g *Gork) AddPeer(p Peer) error {
	if g.HasPeer(p) {
		return ErrPeerExists
	}
	g.Peers = append(g.Peers, p)
	return nil
}

func (g *Gork) AsPeer() Peer {
	k := g.PublicKey()
	return Peer{k}
}

func (g *Gork) MarshalPEM() ([]byte, error) {
	block := &pem.Block{
		Type: "Goracle Private Key",
		Headers: map[string]string{
			"nickname": "charlie",
			"pubkey":   fmt.Sprintf("%x", g.PublicKey().Bytes()),
		},
		Bytes: g.PrivateKey().Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}

func (g *Gork) UnmarshalPEM(b []byte) error {
	block, _ := pem.Decode(b)
	if block == nil {
		return errors.New("bad block")
	}
	privkey := block.Bytes
	pub, err := hex.DecodeString(block.Headers["pubkey"])
	if err != nil {
		return err
	}
	kp := delphi.KeyPair{}
	kp[0] = delphi.Key{}.From(pub)
	kp[1] = delphi.Key{}.From(privkey)
	g.Principal = kp
	return nil
}
