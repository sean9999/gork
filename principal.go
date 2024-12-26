package gork

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/sean9999/go-delphi"
	stablemap "github.com/sean9999/go-stable-map"
)

// KV is a key-value store whose keys are ordered
type KV = stablemap.StableMap[string, string]

func NewKV() *KV {
	return stablemap.New[string, string]()
}

// a Principal is a public/private key-pair with some properties, and knowlege of [Peer]s
type Principal struct {
	delphi.Principal `msgpack:"priv" json:"priv" yaml:"priv"`
	Properties       *KV    `msgpack:"props" json:"props" yaml:"props"`
	Peers            []Peer `msgpack:"peers" json:"peers" yaml:"peers"`
}

// Compose creates a message for a recipient
func (g Principal) Compose(body []byte, headers *KV, recipient Peer) delphi.Message {
	msg := delphi.Message{
		Recipient: recipient.Key,
		Sender:    g.PublicKey(),
		PlainText: body,
	}
	return msg
}

// Art returns the ASCII art representing a public key.
// It can be used for easy visual identification.
func (g Principal) Art() string {
	return g.AsPeer().Art()
}

// func (g Gork) Encrypt(randy io.Reader, msg *delphi.Message, opts any) error {
// 	return msg.Encrypt(randy, g, nil)
// }

// func (g Gork) Decrypt(msg *delphi.Message) error {
// 	return g.Principal.Decrypt(msg *delphi.Message, opts crypto.DecrypterOpts)
// }

// NewPrincipal creates a new [Principal].
func NewPrincipal(randy io.Reader, m map[string]string) Principal {
	prince := delphi.NewPrincipal(randy)
	peers := make([]Peer, 0, 8)
	sm := NewKV()
	sm.Incorporate(m)
	return Principal{*prince, sm, peers}
}

// HasPeer returns true of the Principal has knowlege of that Peer
func (g *Principal) HasPeer(p Peer) bool {
	return slices.Contains(g.Peers, p)
}

var ErrPeerExists = errors.New("peer already exists")

// DropPeer makes a Principal forget a Peer.
func (g *Principal) DropPeer(p Peer) {
	p2 := make([]Peer, 0, len(g.Peers))
	for _, thisPeer := range g.Peers {
		if !thisPeer.Equal(p) {
			p2 = append(p2, thisPeer)
		}
	}
	g.Peers = p2
}

// AddPeer adds a Peer to a Principal's address book.
func (g *Principal) AddPeer(p Peer) error {
	if g.HasPeer(p) {
		return ErrPeerExists
	}
	g.Peers = append(g.Peers, p)
	return nil
}

// AsPeer converts a Principal (public and private key) to a Peer (just public key)
func (g *Principal) AsPeer() Peer {
	k := g.PublicKey()
	return Peer{k, g.Properties}
}

// MarshalPEM marshals a Principal to PEM format
func (g *Principal) MarshalPEM() ([]byte, error) {
	headers := g.Properties.AsMap()
	headers["grip"] = g.AsPeer().Grip()
	headers["pubkey"] = base64.StdEncoding.EncodeToString(g.AsPeer().Bytes())
	block := &pem.Block{
		Type:    "ORACLE PRIVATE KEY",
		Headers: headers,
		Bytes:   g.PrivateKey().Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}

var ErrBadPem = errors.New("malformed pem")
var ErrBadHex = errors.New("bad hex")

// UnmarshalPEM converts a PEM to a Principal
func (g *Principal) UnmarshalPEM(b []byte) error {
	block, _ := pem.Decode(b)
	if block == nil {
		return ErrBadPem
	}
	privkey := block.Bytes
	pub64, exists := block.Headers["pubkey"]
	if !exists {
		return ErrNoPubKey
	}
	pub, err := base64.StdEncoding.DecodeString(pub64)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBadHex, err)
	}
	kp := delphi.KeyPair{}
	kp[0] = delphi.Key{}.From(pub)
	kp[1] = delphi.Key{}.From(privkey)
	g.Principal = kp
	g.Properties = stablemap.From(block.Headers)
	g.Properties.Delete("grip")
	return nil
}
