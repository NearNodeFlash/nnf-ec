package main

import (
	"flag"
	"fmt"

	"stash.us.cray.com/rabsw/nnf-ec/internal/kvstore"
)

func main() {
	var path string
	flag.StringVar(&path, "path", "nnf.db", "the kvstore database to display")
	flag.Parse()

	fmt.Printf("Debug KVStore Tool. Path: '%s'\n", path)
	store, err := kvstore.Open(path, true)
	if err != nil {
		panic(err)
	}

	store.Register([]kvstore.Registry{&debugRegistry{}})

	if err := store.Replay(); err != nil {
		panic(err)
	}
}

type debugRegistry struct{}

func (*debugRegistry) Prefix() string                            { return "" }
func (*debugRegistry) NewReplay(id string) kvstore.ReplayHandler { return &debugReplayHandler{id: id} }

type debugReplayHandler struct {
	id string
}

func (rh *debugReplayHandler) Metadata(data []byte) error {
	fmt.Printf("Running Replay %s:\n", rh.id)
	fmt.Printf("|\tMetadata: %s\n", string(data))
	return nil
}

func (rh *debugReplayHandler) Entry(t uint32, data []byte) error {
	fmt.Printf("|\t\tType: %d Data: %s\n", t, string(data))
	return nil
}

func (rh *debugReplayHandler) Done() error {
	fmt.Printf("|-Replay Done %s\n", rh.id)
	return nil
}
