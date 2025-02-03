package gork

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/sean9999/go-delphi"
	"github.com/sean9999/pear"
)

var ErrNoPrivKey = pear.Defer("no private key")
var ErrNoPubKey = pear.Defer("no public key")

// type propsAndVerity struct {
// 	Props  KV     `yaml:"props,omitempty" json:"props,omitempty" msgpack:"props,omitempty"`
// 	Verity Verity `yaml:"ver" json:"ver" msgpack:"ver"`
// }

// a Config is an object suitable for serializing and storing [Peer]s and key-value pairs
type Config struct {
	readBuf []byte
	Pub     delphi.Key `yaml:"pub" json:"pub" msgpack:"pub"`
	Props   *KV        `yaml:"props,omitempty" json:"props,omitempty" msgpack:"props,omitempty"`
	Peers   *PeerList  `yaml:"peers,omitempty" json:"peers,omitempty" msgpack:"peers,omitempty"`
	//File    afero.File `yaml:"-" json:"-" msgpack:"-"`
	Verity *Verity `yaml:"ver" json:"ver" msgpack:"ver"`
}

func (c Config) Verify(p Principal) (bool, error) {
	dig, err := c.Digest()
	if err != nil {
		return false, err
	}
	return p.Verify(p.PublicKey(), dig, c.Verity.Signature), nil
}

// Hydrate fills a [Config] with information from a [Principal]
func (c *Config) Hydrate(p *Principal) {
	c.Pub = p.PublicKey()
	c.Props = p.Props
	c.Peers = &p.Peers
}

func (c *Config) ensureNonce(randy io.Reader) error {

	if randy == nil {
		return errors.New("nil randomness")
	}

	if c.Verity == nil {
		c.Verity = &Verity{}
	}
	id, err := uuid.NewRandomFromReader(randy)
	if err != nil {
		return pear.Errorf("can't ensure nonce: %w", err)
	}
	c.Verity.Nonce = id[:]
	return nil
}

// type PeerConfig struct {
// 	Pub    delphi.Key `yaml:"pub" json:"pub" msgpack:"pub"`
// 	Props  KV         `yaml:"props,omitempty" json:"props,omitempty" msgpack:"props,omitempty"`
// 	Verity *Verity    `yaml:"ver" json:"ver" msgpack:"ver"`
// }

// Verity is a struct that holds a signature plus some randomness that was involved in calculating the signature.
type Verity struct {
	Nonce     []byte `yaml:"nonce" json:"nonce" msgpack:"nonce"`
	Signature []byte `yaml:"sig" json:"sig" msgpack:"sig"`
}

func (v *Verity) MarshalJSON() ([]byte, error) {
	nonce := hex.EncodeToString(v.Nonce)
	sig := hex.EncodeToString(v.Signature)
	m := map[string]string{
		"nonce": nonce,
		"sig":   sig,
	}
	return json.Marshal(m)
}

func (v *Verity) UnmarshalJSON(b []byte) error {
	var m map[string]string

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	nonce, err := hex.DecodeString(m["nonce"])
	if err != nil {
		return err
	}
	sig, err := hex.DecodeString(m["sig"])
	if err != nil {
		return err
	}

	v.Nonce = nonce
	v.Signature = sig
	return nil
}

// func (v Verity) MarshalJSON() ([]byte, error) {
// 	str := fmt.Sprintf("%x.%x", v.Nonce, v.Signature)
// 	return []byte(str), nil
// }

// func (v *Verity) UnmarshalJSON(b []byte) error {
// 	slug := strings.Split(string(b), ".")
// 	if len(slug) != 2 {
// 		return errors.New("bad slug length")
// 	}
// 	nonce, err := hex.DecodeString(slug[0])
// 	if err != nil {
// 		return err
// 	}
// 	sig, err := hex.DecodeString(slug[1])
// 	if err != nil {
// 		return err
// 	}
// 	v.Nonce = nonce
// 	v.Signature = sig
// 	return nil
// }

// produce a digest, for signing
func (c *Config) Digest() (digest []byte, err error) {

	if len(c.Verity.Nonce) == 0 {
		return nil, pear.New("nil nonce")
	}

	fields := [3][]byte{}
	//	0 : pub key
	//	1 : props
	//	3 : nonce

	props, err := c.Props.MarshalJSON()
	if err != nil {
		return nil, err
	}

	fields[0] = c.Pub.Bytes()
	fields[1] = props

	if c.Verity == nil {
		c.Verity = &Verity{}
	}

	fields[2] = c.Verity.Nonce

	for _, field := range fields {
		digest = append(digest, field...)
	}
	return digest, nil

}

// func (c *Config) WithDescriptor(f afero.File) {
// 	io.Copy(c, f)
// 	c.File = f
// }

// copy values from c to d
// func (c *Config) cloneInto(d *Config) {
// 	*d = *c
// }

func (c *Config) Write(b []byte) (int, error) {
	//d := new(Config)
	err := json.Unmarshal(b, c)
	if err != nil {
		return 0, pear.Errorf("could not write to config. %w", err)
	}
	return len(b), io.EOF
}

func (c *Config) Read(b []byte) (int, error) {

	if c.readBuf == nil {
		buf, err := json.MarshalIndent(c, "", "\t")
		if err != nil {
			return 0, err
		}
		buf = append(buf, []byte("\n")...)
		c.readBuf = buf
	}

	if len(c.readBuf) == 0 {
		return 0, io.EOF
	}

	if len(b) == 0 {
		return 0, nil
	}

	i := copy(b, c.readBuf)
	c.readBuf = c.readBuf[i:]

	var err error
	if len(c.readBuf) == 0 {
		err = io.EOF
	}
	return i, err
}

// func (c Config) MarshalJSON() ([]byte, error) {
// 	return json.MarshalIndent(c, "", "\t")
// }

// func (c *Config) UnmarshalJSON(b []byte) error {
// 	var m map[string]any

// 	c.Props = NewKV()
// 	c.Peers = map[delphi.Key]KV{}
// 	err := json.Unmarshal(b, &m)
// 	if err != nil {
// 		return pear.Errorf("could not unmarshal config: %w", err)
// 	}
// 	for k, v := range m {
// 		switch k {
// 		default:
// 			//c.Props.Set(k, v.(string))
// 		case "pub":
// 			c.Pub = delphi.KeyFromHex(v.(string))
// 		case "peers":
// 			for pubkey, peer := range v.(map[delphi.Key]KV) {
// 				c.Peers[pubkey] = peer
// 			}
// 		}
// 	}
// 	return nil
// }

func peerToConfig(p Peer) Config {
	c := Config{
		Pub:   p.Key,
		Props: p.Properties,
	}
	return c
}

func PrincipalToConfig(p Principal) Config {
	conf := peerToConfig(p.AsPeer())
	//conf.Peers = p.Peers
	return conf
}

func configToPeer(c Config) (*Peer, error) {
	if c.Pub.IsZero() {
		return nil, ErrNoPubKey
	}
	p := Peer{
		Key:        delphi.Key{}.From(c.Pub.Bytes()),
		Properties: c.Props,
	}
	return &p, nil
}

func NewConfig() *Config {

	peerlist := make(PeerList, 0)
	c := Config{
		Pub:   delphi.Key{},
		Props: NewKV(),
		Peers: &peerlist,
	}
	return &c
}

// func configToPrincipal(c Config) (*Principal, error) {

// 	// if c.Priv.IsZero() {
// 	// 	return nil, ErrNoPrivKey
// 	// }
// 	dp := delphi.Principal{}.From(c.Priv.Bytes())

// 	//	peers
// 	peers := make([]Peer, len(c.Peers))
// 	for i, peerConf := range c.Peers {
// 		peer, err := configToPeer(peerConf)
// 		if err != nil {
// 			return nil, fmt.Errorf("could not convert config to principal: %w", err)
// 		}
// 		peers[i] = *peer
// 	}

// 	p := Principal{
// 		Principal:  dp,
// 		Properties: c.Props,
// 		Peers:      peers,
// 	}
// 	return &p, nil
// }
