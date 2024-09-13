package log

import (
	"github.com/hashicorp/raft"
)

// A Config is the configuration for the log.
type Config struct {
	Raft struct {
		raft.Config              // configure the raft configuration
		StreamLayer *StreamLayer // configure the stream layer
		Bootstrap   bool         // configure the bootstrap flag
	}
	Segment struct {
		MaxStoreBytes uint64 // configure the maximum size of the store
		MaxIndexBytes uint64 // configure the maximum size of the index
		InitialOffset uint64 // configure the initial offset for the first segment
	}
}
