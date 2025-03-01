package gork

import (
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"io"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/pear"
)

// a Principal is a public/private key-pair with some properties, and knowlege of [Peer]s
type Principal struct {
	delphi.Principal `msgpack:"priv" json:"priv" yaml:"priv"`
	Props            KV        `msgpack:"props" json:"props" yaml:"props"`
	Peers            PeerMap   `msgpack:"peers" json:"peers" yaml:"peers"`
	randomness       io.Reader `msgpack:"-" json:"-" yaml:"-"`
}

// Compose creates a message for a recipient. It's syntactic sugar for [delphi.NewMessage]
func (g *Principal) Compose(body []byte, headers delphi.KV, recipient Peer) *delphi.Message {
	msg := delphi.NewMessage(g.randomness, body)
	if headers != nil {
		msg.Headers = headers
	} else {
		msg.Headers = make(delphi.KV)
	}
	msg.Recipient = recipient.Key
	msg.Sender = g.PublicKey()
	return msg
}

// Art returns the ASCII art representing a public key.
// It can be used for easy visual identification.
func (g *Principal) Art() string {
	return g.AsPeer().Art()
}

// func (g Gork) Encrypt(randy io.Reader, msg *delphi.Message, opts any) error {
// 	return msg.Encrypt(randy, g, nil)
// }

// func (g Gork) Decrypt(msg *delphi.Message) error {
// 	return g.Principal.Decrypt(msg *delphi.Message, opts crypto.DecrypterOpts)
// }

// func mapToPairs[K comparable, V any](m map[K]V) []omap.Pair[K, V] {
// 	pairs := make([]omap.Pair[K,V],0,len(m))

// 	for k,v := range m {
// 		pair := omap.Pair[K,V]{
// 			Key: k,
// 			Value: v,
// 		}
// 		//pair := omap.Pair[K,V]{k,v}
// 	}

// }

type option func(*Principal)

func WithRand(randy io.Reader) option {
	return func(p *Principal) {
		p.randomness = randy
		p.Principal = *delphi.NewPrincipal(randy)
	}
}

func WithProps(m map[string]string) option {
	return func(p *Principal) {
		p.Props = m
	}
}

func WithConfig(r io.Reader) option {
	return func(p *Principal) {
		err := p.Load(r)
		if err != nil {
			panic(err)
		}
		// confBytes, err := io.ReadAll(r)
		// if err != nil {
		// 	panic(err)
		// }
		// conf := new(Config)
		// err = json.Unmarshal(confBytes, conf)
		// if err != nil {
		// 	panic(err)
		// }
		// p.Peers = conf.Peers
		// maps.Copy(p.Props, conf.Props)
	}
}

func ConstructPrincipal(options ...option) *Principal {
	p := &Principal{
		Props:      make(KV),
		Peers:      make(PeerMap),
		randomness: rand.Reader,
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

// NewPrincipal creates a new [Principal].
// func NewPrincipal(randy io.Reader, m map[string]string, prov io.ReadWriteCloser) *Principal {
// 	prince := delphi.NewPrincipal(randy) // random private key
// 	peers := make(PeerMap, 0)
// 	props := NewKV()
// 	king := &Principal{*prince, props, peers, randy}

// 	if prov != nil {
// 		err := king.Load(prov)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	return king
// }

func (g *Principal) WithRand(randy io.Reader) {
	g.randomness = randy
}

// func (f FileBasedConfigProvider) Get() (*Config, error) {
// 	fd, err := f.openForReading()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer fd.Close()
// 	fileBytes, err := io.ReadAll(fd)
// 	if err != nil {
// 		return nil, err
// 	}
// 	conf := NewConfig()

// 	err = json.Unmarshal(fileBytes, conf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conf, nil
// }

// func (f FileBasedConfigProvider) Set(c *Config) error {
// 	if c == nil {
// 		return errors.New("nil config")
// 	}
// 	fd, err := f.openForWriting()
// 	if err != nil {
// 		return err
// 	}
// 	_, err = io.Copy(fd, c)
// 	return err
// }

// func (g *Principal) WithConfigFile(filesytem afero.Fs, fileName string) error {
// 	prov := FileBasedConfigProvider{
// 		Fs:   filesytem,
// 		Name: fileName,
// 	}
// 	return g.WithConfigProvider(&prov)
// }

// func (g *Principal) WithConfigProvider(prov ConfigProvider) error {

// 	if g.ConfigProvider != nil {
// 		return errors.New("config provider already exists")
// 	}

// 	g.ConfigProvider = prov
// 	conf, err := prov.Get()
// 	if err != nil {
// 		return pear.Errorf("could not get config file. %w", err)
// 	}
// 	return g.LoadConfig(conf)
// }

// load a config file and attach data to a [Principal]
// func (g *Principal) LoadConfig(c *Config) error {
// 	//	TODO: we could verify that pubkeys match

// 	if c.Pub.IsZero() {
// 		return errors.New("zero public key")
// 	}

// 	if c.Peers != nil {
// 		g.Peers = *c.Peers
// 	}
// 	g.Props = c.Props
// 	return nil
// }

// Save writes the Principal's Peers and custom properties to a config file
// func (g *Principal) Save(prov ConfigProvider) error {
// 	if g == nil {
// 		return pear.New("nil principal")
// 	}
// 	if prov == nil && g.ConfigProvider != nil {
// 		prov = g.ConfigProvider
// 	}
// 	if prov == nil {
// 		return pear.New("nil config provider")
// 	}
// 	conf := g.Export()
// 	return prov.Set(conf)
// }

// // HasPeer returns true if the Principal has knowlege of that Peer
// func (g *Principal) HasPeer(p Peer) bool {
// 	for _, peer := range g.Peers {
// 		if peer.Equal(p) {
// 			return true
// 		}
// 	}
// 	return false
// }

var ErrPeerExists = pear.Defer("peer already exists")

// DropPeer makes a Principal forget a Peer.
func (g *Principal) DropPeer(p Peer) {
	delete(g.Peers, p.Key)
}

// AddPeer adds a Peer to a Principal's address book.
func (g *Principal) AddPeer(p Peer) error {
	if g.HasPeer(p) {
		return ErrPeerExists
	}
	g.Peers[p.Key] = p.Properties
	return nil
}

// func (g *Principal) WithConfigProvider(rw io.ReadWriteCloser) {
// 	g.configProvider = rw
// }

// func (g *Principal) Save(w io.Writer) error {
// 	if w == nil {
// 		w = g.configProvider
// 	}
// 	if w == nil {
// 		return errors.New("nil config provider")
// 	}
// 	_, err := io.Copy(w, g)
// 	return err
// }

func (g *Principal) HasPeer(p Peer) bool {
	_, exists := g.Peers[p.Key]
	return exists
}

func (g *Principal) Nickname() string {
	return g.AsPeer().Nickname()
}

// AsPeer converts a Principal (public and private key) to a Peer (just public key)
func (g *Principal) AsPeer() Peer {
	k := g.PublicKey()
	return Peer{k, g.Props}
}

// MarshalPEM marshals a Principal to PEM format
func (g *Principal) MarshalPEM() ([]byte, error) {
	// headers := make(map[string]string, g.Props.Len())
	// for pair := g.Props.Oldest(); pair != nil; pair = pair.Next() {
	// 	k, v := pair.Key, pair.Value
	// 	headers[k] = v
	// }

	headers := make(map[string]string, 3)
	//headers["pubkey"] = g.AsPeer().ToHex()
	headers["grip"] = g.AsPeer().Grip()
	headers["nick"] = g.AsPeer().Nickname()

	block := &pem.Block{
		Type:    "ORACLE PRIVATE KEY",
		Headers: headers,
		Bytes:   g.Bytes(),
	}
	return pem.EncodeToMemory(block), nil
}

var ErrBadPem = errors.New("malformed pem")
var ErrBadHex = errors.New("bad hex")

// UnmarshalPEM converts a PEM to a Principal
func (g *Principal) UnmarshalPEM(b []byte) error {
	block, _ := pem.Decode(b)
	if block == nil {
		return fmt.Errorf("could not decode pem. %w", ErrBadPem)
	}
	privkey := block.Bytes
	// pub64, exists := block.Headers["pubkey"]
	// if !exists {
	// 	return ErrNoPubKey.Throw(1)
	// }
	// pub, err := base64.StdEncoding.DecodeString(pub64)
	// if err != nil {
	// 	return pear.Errorf("%w: %w", ErrBadHex, err)
	// }

	// sm := NewKV()
	// incorporate(sm, block.Headers)

	prince, err := delphi.Principal{}.From(privkey)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBadPem, err)
	}
	g.Principal = prince
	if g.Props == nil {
		g.Props = NewKV()
	}
	return nil
}

func (g *Principal) FromPem(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return g.UnmarshalPEM(b)
}

func (g *Principal) FromBin(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return g.UnmarshalBinary(b)
}

func (g *Principal) ToBin() []byte {
	return g.Principal.Bytes()
}

// PrincipalFrom assumes binary format, but maybe it should assume PEM
func PrincipalFrom(r io.Reader) (*Principal, error) {
	p := new(Principal)
	err := p.FromBin(r)
	return p, err
}
