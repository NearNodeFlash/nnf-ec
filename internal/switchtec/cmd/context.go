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

package cmd

type DebugLevel int

type LogLevel int

// Context provides the CLI context global to all commands
type Context struct {
	DebugLevel DebugLevel
	LogLevel   LogLevel
}

const (
	// Disabled defines the debug/logging level where all logging is disabled.
	Disabled = iota

	// Debug defines the debug/log level of basic debug information within the package.
	Debug

	// Info defines the debug/log level of informative or noteworthy information
	Info

	// Warn defines the debug/log level of warnings and other things that should be investigated in the package
	Warn

	// Error defines the debug/log level of errors within the package
	Error

	// Fatal defines the debug/log level of fatal events within the package.
	Fatal

	// Panic defines the debug/log level of panics within the package. These are unrecoverable.
	Panic
)

var (
	globalDebugLevel DebugLevel = Disabled
	globalLogLevel   LogLevel   = Disabled
)

// ApplyContext will load the context into the cmd package
func ApplyContext(ctx Context) {
	globalDebugLevel = ctx.DebugLevel
	globalLogLevel = ctx.LogLevel
}
