package gork

import "github.com/sean9999/go-delphi"

type KV = delphi.KV

var NewKV = delphi.NewKV

// type KV = stablemap.StableMap[string, string]

// func NewKV() *KV {
// 	sm := stablemap.New[string, string]()
// 	kv := KV(*sm)
// 	return &kv
// }
