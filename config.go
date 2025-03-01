package gork

import (
	"encoding/json"
	"io"

	"github.com/sean9999/go-delphi"
)

type Config struct {
	Pub    delphi.Key        `json:"pub"`
	Props  KV                `json:"props"`
	Peers  PeerMap           `json:"peers"`
	Verity map[string]string `json:"ver,omitempty"`
}

func (p *Principal) Save(w io.Writer) error {

	defer func() {
		if wc, ok := w.(io.Closer); ok {
			wc.Close()
		}
	}()

	conf := Config{
		Pub:   p.PublicKey(),
		Props: p.Props,
		Peers: p.Peers,
	}
	jsonBytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func (p *Principal) Load(r io.Reader) error {

	defer func() {
		if rc, ok := r.(io.Closer); ok {
			rc.Close()
		}
	}()

	confBytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	conf := NewConfig()
	err = json.Unmarshal(confBytes, &conf)
	if err != nil {
		return err
	}
	//	TODO: compare pubkeys
	p.Props = conf.Props
	p.Peers = conf.Peers
	return nil
}

func NewConfig() Config {
	return Config{
		Pub:    delphi.Key{},
		Props:  make(KV),
		Peers:  make(PeerMap),
		Verity: make(map[string]string),
	}
}
