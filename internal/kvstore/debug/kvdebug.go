/*
 * Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"

	"github.com/NearNodeFlash/nnf-ec/internal/kvstore"
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
