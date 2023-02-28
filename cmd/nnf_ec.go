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

	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	nnf "github.com/NearNodeFlash/nnf-ec/pkg"
	ec "github.com/NearNodeFlash/nnf-ec/pkg/ec"
)

func main() {

	// I would love for this to read a little better and FORCE usage of the command line flags,
	// not make it optional. EC does not need an options interface, it should do everything
	// when Run() is called (if there is no harm to running flag.Parse() twice?).
	// Similarly, the NNF Controller should hide the flags it needs in the call to New()

	nnfOpts := nnf.BindFlags(flag.CommandLine)
	ecOpts := ec.BindFlags(flag.CommandLine)

	zapOpts := zap.Options{
		Development: true,
		Level:       zapcore.Level(-3),
	}
	zapOpts.BindFlags(flag.CommandLine)

	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&zapOpts))

	c := nnf.NewController(nnfOpts).WithLogger(logger)

	c.Init(ecOpts)
	c.Run()
}
