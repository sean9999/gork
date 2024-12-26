package gork

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/sean9999/go-delphi"
	stablemap "github.com/sean9999/go-stable-map"
)

var ErrNoPrivKey = errors.New("no private key")
var ErrNoPubKey = errors.New("no public key")

// a Config is an object suitable for serializing and storing [Peer]s
type Config struct {
	//Priv  delphi.Key                           `yaml:"priv" json:"priv" msgpack:"priv"`
	Pub   delphi.Key                           `yaml:"pub" json:"pub" msgpack:"pub"`
	Props *stablemap.StableMap[string, string] `yaml:"props" json:"props" msgpack:"props"`
	Peers []Config                             `yaml:"peers" json:"peers" msgpack:"peers"`
}

func (c *Config) Write(b []byte) (int, error) {
	d := new(Config)
	err := d.UnmarshalJSON(b)
	if err != nil {
		return 0, fmt.Errorf("could not write to config. %w", err)
	}
	c = d
	return len(b), io.EOF
}

func (c *Config) Read(b []byte) (int, error) {
	var err error = nil
	jsonBytes, err := c.MarshalJSON()
	if err != nil {
		return 0, fmt.Errorf("could not read from config. %w", err)
	}
	i := copy(b, jsonBytes)
	if i == len(jsonBytes) {
		err = io.EOF
	}
	return i, err
}

func (c Config) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(c.Peers)+2)
	// if !c.Priv.IsZero() {
	// 	m["priv"] = fmt.Sprintf("%x", c.Priv.Bytes())
	// }
	m["pub"] = fmt.Sprintf("%x", c.Pub.Bytes())
	for k, v := range c.Props.Entries() {
		m[k] = v
	}
	if len(c.Peers) > 0 {
		m["peers"] = c.Peers
	}
	return json.Marshal(m)
}

func (c *Config) UnmarshalJSON(b []byte) error {
	var m map[string]any
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	for k, v := range m {
		switch k {
		// case "priv":
		// 	c.Priv = delphi.KeyFromHex(v.(string))
		case "pub":
			c.Pub = delphi.KeyFromHex(v.(string))
		case "peers":
			for i, conf := range v.([]Config) {
				c.Peers[i] = conf
			}
		}
	}
	return nil
}

func peerToConfig(p Peer) Config {
	c := Config{
		Pub:   p.Key,
		Props: p.Properties,
	}
	return c
}

func PrincipalToConfig(p Principal) Config {
	conf := peerToConfig(p.AsPeer())
	//conf.Priv = p.PrivateKey()
	peers := make([]Config, len(p.Peers))
	for i, peer := range p.Peers {
		peers[i] = peerToConfig(peer)
	}
	conf.Peers = peers
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
