package gork

import (
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/sean9999/go-delphi"
	stablemap "github.com/sean9999/go-stable-map"
)

type KV = stablemap.StableMap[string, string]

func NewKV() *KV {
	return stablemap.New[string, string]()
}

type Principal struct {
	delphi.Principal
	Properties *KV
	Peers      []Peer
}

func (g Principal) Compose(body []byte, headers *KV, recipient Peer) delphi.Message {
	msg := delphi.Message{
		Recipient: recipient.Key,
		Sender:    g.PublicKey(),
		PlainText: body,
	}
	return msg
}

func (g Principal) Art() string {
	return g.AsPeer().Art()
}

// func (g Gork) Encrypt(randy io.Reader, msg *delphi.Message, opts any) error {
// 	return msg.Encrypt(randy, g, nil)
// }

// func (g Gork) Decrypt(msg *delphi.Message) error {
// 	return g.Principal.Decrypt(msg *delphi.Message, opts crypto.DecrypterOpts)
// }

func NewPrincipal(randy io.Reader, m map[string]string) Principal {
	prince := delphi.NewPrincipal(randy)
	peers := make([]Peer, 0, 8)
	sm := NewKV()
	sm.Incorporate(m)
	return Principal{*prince, sm, peers}
}

func (g *Principal) HasPeer(p Peer) bool {
	return slices.Contains(g.Peers, p)
}

var ErrPeerExists = errors.New("peer already exists")

func (g *Principal) DropPeer(p Peer) {
	p2 := make([]Peer, 0, len(g.Peers))
	for _, thisPeer := range g.Peers {
		if !thisPeer.Equal(p) {
			p2 = append(p2, thisPeer)
		}
	}
	g.Peers = p2
}

func (g *Principal) AddPeer(p Peer) error {
	if g.HasPeer(p) {
		return ErrPeerExists
	}
	g.Peers = append(g.Peers, p)
	return nil
}

func (g *Principal) AsPeer() Peer {
	k := g.PublicKey()
	return Peer{k, g.Properties}
}

func (g *Principal) MarshalPEM() ([]byte, error) {
	headers := g.Properties.AsMap()
	headers["grip"] = g.AsPeer().Grip()
	headers["pubkey"] = fmt.Sprintf("%x", g.PublicKey().Bytes())
	block := &pem.Block{
		Type:    "ORACLE PRIVATE KEY",
		Headers: headers,
		Bytes:   g.PrivateKey().Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}

var ErrNoPubkey = errors.New("no pub key")
var ErrBadPem = errors.New("malformed pem")
var ErrBadHex = errors.New("bad hex")

func (g *Principal) UnmarshalPEM(b []byte) error {
	block, _ := pem.Decode(b)
	if block == nil {
		return ErrBadPem
	}
	privkey := block.Bytes
	hexPub, exists := block.Headers["pubkey"]
	if !exists {
		return ErrNoPubkey
	}
	pub, err := hex.DecodeString(hexPub)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBadHex, err)
	}
	kp := delphi.KeyPair{}
	kp[0] = delphi.Key{}.From(pub)
	kp[1] = delphi.Key{}.From(privkey)
	g.Principal = kp
	g.Properties = stablemap.From(block.Headers)
	return nil
}
