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

package remote

import (
	"flag"

	ec "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/ec"
	nnf "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-nnf"
	server "github.hpe.com/hpe/hpc-rabsw-nnf-ec/pkg/manager-server"
)

type Options struct{}

func BindFlags(fs *flag.FlagSet) *Options {
	return &Options{}
}

func RunController() error {

	opts := BindFlags(flag.CommandLine)

	flag.Parse()

	c := &ec.Controller{
		Name: "Near Node Flash Server",
		Routers: []ec.Router{
			nnf.NewDefaultApiRouter(
				nnf.NewDefaultApiService(
					NewDefaultServerStorageService(opts),
				), nil),
		},
	}

	err := c.Init(&ec.Options{
		Http:    true,
		Port:    server.RemoteStorageServicePort,
		Log:     true,
		Verbose: true,
	})

	if err != nil {
		return err
	}

	c.Run()

	return nil
}
