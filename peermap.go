package gork

import (
	"encoding/json"

	"github.com/sean9999/go-delphi"
)

type PeerMap map[delphi.Key]KV

func (pm PeerMap) MarhshalJSON() ([]byte, error) {
	m := map[string]KV{}
	for key, props := range pm {
		m[key.ToHex()] = props
	}
	return json.Marshal(m)
}

func (pmPtr *PeerMap) UnmarshalJSON(b []byte) error {
	var pm PeerMap
	if pmPtr == nil {
		pm = make(PeerMap)
		pmPtr = &pm
	} else {
		pm = *pmPtr
	}
	var m map[string]KV
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	for hexKey, props := range m {
		pm[delphi.KeyFromHex(hexKey)] = props
	}
	return nil
}

func (pm PeerMap) Get(k delphi.Key) Peer {
	p := NewPeer(k.Bytes())
	p.Properties = pm[k]
	return p
}

func (pm PeerMap) Add(p Peer) {
	pm[p.Key] = p.Properties
}
