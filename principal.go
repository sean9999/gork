package gork

import (
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/pear"
	"github.com/spf13/afero"
	omap "github.com/wk8/go-ordered-map/v2"
)

// KV is a key-value store whose keys are ordered, offering deterministic serialization
type KV = omap.OrderedMap[string, string]

func NewKV() *KV {
	return omap.New[string, string]()
}

// a Principal is a public/private key-pair with some properties, and knowlege of [Peer]s
type Principal struct {
	delphi.Principal `msgpack:"priv" json:"priv" yaml:"priv"`
	Props            *KV            `msgpack:"props" json:"props" yaml:"props"`
	Peers            PeerList       `msgpack:"peers" json:"peers" yaml:"peers"`
	randomness       io.Reader      `msgpack:"-" json:"-" yaml:"-"`
	ConfigProvider   ConfigProvider `msgpack:"-" json:"-" yaml:"-"`
}

// Export produces a *Config from a *Principal
func (p *Principal) Export() *Config {
	conf := NewConfig()
	//conf.Hydrate(p)
	p.SignConfig(conf)
	return conf
}

// SignConfig signs a config calculating a digest, adding a nonce, and producing a signature
func (g *Principal) SignConfig(conf *Config) error {

	conf.Hydrate(g)

	err := conf.ensureNonce(g.randomness)

	if err != nil {
		return err
	}

	digest, err := conf.Digest()
	if err != nil {
		return err
	}

	sig, err := g.Principal.Sign(nil, digest, nil)
	if err != nil {
		return err
	}

	conf.Verity.Signature = sig
	return nil
}

func (g *Principal) VerifyConfig(c *Config) error {

	pub := g.PublicKey()
	dig, err := c.Digest()
	if err != nil {
		return err
	}
	sig := c.Verity.Signature
	ok := g.Principal.Verify(pub, dig, sig)
	if !ok {
		return errors.New("verification failed")
	}
	return nil
}

// Compose creates a message for a recipient. It's syntactic sugar for [delphi.NewMessage]
func (g *Principal) Compose(body []byte, headers *delphi.KV, recipient Peer) *delphi.Message {
	msg := delphi.NewMessage(g.randomness, body)
	if headers != nil {
		msg.Headers = headers
	} else {
		msg.Headers = NewKV()
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

func incorporate(omap *KV, m map[string]string) {
	for k, v := range m {
		omap.Set(k, v)
	}
}

// NewPrincipal creates a new [Principal].
func NewPrincipal(randy io.Reader, m map[string]string, prov ConfigProvider) Principal {
	prince := delphi.NewPrincipal(randy)
	peers := make(PeerList, 0)
	sm := NewKV()
	king := Principal{*prince, sm, peers, randy, prov}
	err := king.ensureGrip()
	if err != nil {
		panic(err)
	}
	incorporate(sm, m)
	return king
}

func (g *Principal) ensureGrip() error {
	g.Props.Set("grip", g.AsPeer().Grip())
	g.Props.MoveToFront("grip")
	return nil
}

func (g *Principal) WithRand(randy io.Reader) {
	g.randomness = randy
}

type FileBasedConfigProvider struct {
	Fs   afero.Fs
	Name string
}

func (f FileBasedConfigProvider) openForReading() (afero.File, error) {
	return f.Fs.Open(f.Name)
}

func (f FileBasedConfigProvider) openForWriting() (afero.File, error) {
	return f.Fs.OpenFile(f.Name, os.O_RDWR|os.O_TRUNC, 0640)
}

func (f FileBasedConfigProvider) Get() (*Config, error) {
	fd, err := f.openForReading()
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	fileBytes, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	conf := new(Config)
	err = json.Unmarshal(fileBytes, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (f FileBasedConfigProvider) Set(c *Config) error {
	if c == nil {
		return errors.New("nil config")
	}
	fd, err := f.openForWriting()
	if err != nil {
		return err
	}
	_, err = io.Copy(fd, c)
	return err
}

func (g *Principal) WithConfigFile(filesytem afero.Fs, fileName string) error {
	prov := FileBasedConfigProvider{
		Fs:   filesytem,
		Name: fileName,
	}
	g.ConfigProvider = prov
	conf, err := prov.Get()
	if err != nil {
		return pear.Errorf("could not get config file. %w", err)
	}
	return g.LoadConfig(conf)
}

func (g *Principal) WithConfigProvider(prov ConfigProvider) error {
	g.ConfigProvider = prov
	conf, err := prov.Get()
	if err != nil {
		return pear.Errorf("could not get config file. %w", err)
	}
	return g.LoadConfig(conf)
}

// load a config file and attach data to a [Principal]
func (g *Principal) LoadConfig(c *Config) error {
	//	TODO: we could verify that pubkeys match

	g.Peers = *c.Peers
	g.Props = c.Props
	return nil
}

// Save writes the Principal's Peers and custom properties to a config file
func (g *Principal) Save(prov ConfigProvider) error {
	if prov == nil {
		return pear.New("nil config provider")
	}
	if g == nil {
		return pear.New("nil principal")
	}
	conf := g.Export()
	return prov.Set(conf)
}

// HasPeer returns true if the Principal has knowlege of that Peer
func (g *Principal) HasPeer(p Peer) bool {
	for _, peer := range g.Peers {
		if peer.Equal(p) {
			return true
		}
	}
	return false
}

var ErrPeerExists = pear.Defer("peer already exists")

// DropPeer makes a Principal forget a Peer.
func (g *Principal) DropPeer(p Peer) {
	for i, thisPeer := range g.Peers {
		if thisPeer.Equal(p) {
			g.Peers = append(g.Peers[i+1:], g.Peers[:i]...)
			return
		}
	}
}

// AddPeer adds a Peer to a Principal's address book.
func (g *Principal) AddPeer(p Peer) error {
	if g.HasPeer(p) {
		return ErrPeerExists
	}
	g.Peers = append(g.Peers, p)
	return nil
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
	g.Props = NewKV()
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
