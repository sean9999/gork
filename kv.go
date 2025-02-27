package gork

import (
	"iter"
	"slices"
)

// a KV is a simple map with some super powers useful to us
type KV map[string]string

func NewKV() KV {
	return make(KV)
}

// WithInferred adds inferred keys "nick" and "grip"
func (kv KV) WithInferred(p Peer) KV {
	kv["nick"] = p.Nickname()
	kv["grip"] = p.Grip()
	return kv
}

func (kv KV) WithoutInferred() KV {
	delete(kv, "nick")
	delete(kv, "grip")
	return kv
}

func (kv KV) LexicalOrder() iter.Seq2[string, string] {
	keys := make([]string, 0, len(kv))
	for k, _ := range kv {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return func(yield func(string, string) bool) {
		for _, k := range keys {
			v := kv[k]
			if !yield(k, v) {
				return
			}
		}
	}
}
