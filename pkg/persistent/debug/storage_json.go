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
	"encoding/json"
	"io/ioutil"

	"github.com/NearNodeFlash/nnf-ec/pkg/persistent"
)

func NewJsonFilePersistentStorageProvider(filename string) persistent.PersistentStorageProvider {
	return &jsonFilePersisentStorageProvider{filename: filename}
}

type jsonFilePersisentStorageProvider struct {
	filename string
}

func (p *jsonFilePersisentStorageProvider) NewPersistentStorageInterface(name string, readOnly bool) (persistent.PersistentStorageApi, error) {
	content, err := ioutil.ReadFile(p.filename)
	if err != nil {
		return nil, err
	}

	var payload map[string]map[string]string
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, err
	}

	return &jsonPersistentStorageInterface{data: payload[name]}, nil
}

type jsonPersistentStorageInterface struct {
	data map[string]string
}

func (psi *jsonPersistentStorageInterface) View(fn func(persistent.PersistentStorageTransactionApi) error) error {
	return fn(persistent.NewBase64PersistentStorageTransaction(psi.data))
}

func (*jsonPersistentStorageInterface) Update(func(txn persistent.PersistentStorageTransactionApi) error) error {
	panic("unimplemented")
}

func (*jsonPersistentStorageInterface) Delete(key string) error {
	panic("unimplemented")
}

func (*jsonPersistentStorageInterface) Close() error {
	panic("unimplemented")
}




