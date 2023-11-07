package model

import "sync"

// todo delete this wrapper ( need to fix tests)
type SourceHashes struct {
	Hashes *sync.Map `json:"hashes"`
}
