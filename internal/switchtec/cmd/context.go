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
