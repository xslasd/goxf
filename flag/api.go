package flag

import (
	"context"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	globalMu        sync.RWMutex
	globalFlagSet   = NewFlagSet("goxf")
	globalRootCmd   = &cobra.Command{Use: "goxf"}
	globalCmdInit   sync.Once
)

// RootCommand returns the global root cobra.Command.
func RootCommand() *cobra.Command {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalRootCmd
}

// SetRootCommand sets the global root cobra.Command.
func SetRootCommand(cmd *cobra.Command) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalRootCmd = cmd
}

// GlobalFlagSet returns the global default FlagSet.
func GlobalFlagSet() *FlagSet {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalFlagSet
}

// Reset resets the global flag set and root command (useful in tests or fresh inits).
func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalFlagSet = NewFlagSet("goxf")
	globalRootCmd = &cobra.Command{Use: "goxf"}
}

// Register registers one or more Flag items to the global FlagSet.
func Register(fs ...Flag) {
	globalFlagSet.Register(fs...)
}

// BindStruct automatically binds struct fields with tags to the global FlagSet.
func BindStruct(v interface{}) error {
	return globalFlagSet.BindStruct(v)
}

// AddCommand adds one or more subcommands to the global root command.
func AddCommand(cmds ...*cobra.Command) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalRootCmd.AddCommand(cmds...)
}

// Parse parses the global flags using os.Args[1:] or provided arguments.
func Parse(args ...string) error {
	return globalFlagSet.Parse(args...)
}

// ParseArgs parses the global flags with specified arguments.
func ParseArgs(args []string) error {
	return globalFlagSet.ParseArgs(args)
}

// Execute prepares registered flags and executes the global RootCommand.
func Execute() error {
	prepareRootCommand()
	return globalRootCmd.Execute()
}

// ExecuteContext prepares registered flags and executes the global RootCommand with context.
func ExecuteContext(ctx context.Context) error {
	prepareRootCommand()
	return globalRootCmd.ExecuteContext(ctx)
}

func prepareRootCommand() {
	globalCmdInit.Do(func() {
		globalFlagSet.applyAll()
		globalRootCmd.Flags().AddFlagSet(globalFlagSet.Flags())
		globalRootCmd.PersistentFlags().AddFlagSet(globalFlagSet.PersistentFlags())
	})
}

// Lookup finds a flag by name in the global FlagSet.
func Lookup(name string) *pflag.Flag {
	return globalFlagSet.Lookup(name)
}

// Changed checks if a flag was explicitly set on the CLI.
func Changed(name string) bool {
	return globalFlagSet.Changed(name)
}

// Has checks if a flag is defined in the global FlagSet.
func Has(name string) bool {
	return globalFlagSet.Has(name)
}

// PrintDefaults prints the default values for all registered flags.
func PrintDefaults() {
	globalFlagSet.PrintDefaults()
}

// --- Getter Methods ---

func GetString(name string) (string, error) { return globalFlagSet.GetString(name) }
func String(name string) string             { return globalFlagSet.String(name) }

func GetBool(name string) (bool, error) { return globalFlagSet.GetBool(name) }
func Bool(name string) bool             { return globalFlagSet.Bool(name) }

func GetInt(name string) (int, error) { return globalFlagSet.GetInt(name) }
func Int(name string) int             { return globalFlagSet.Int(name) }

func GetInt64(name string) (int64, error) { return globalFlagSet.GetInt64(name) }
func Int64(name string) int64             { return globalFlagSet.Int64(name) }

func GetUint(name string) (uint, error) { return globalFlagSet.GetUint(name) }
func Uint(name string) uint             { return globalFlagSet.Uint(name) }

func GetUint64(name string) (uint64, error) { return globalFlagSet.GetUint64(name) }
func Uint64(name string) uint64             { return globalFlagSet.Uint64(name) }

func GetFloat64(name string) (float64, error) { return globalFlagSet.GetFloat64(name) }
func Float64(name string) float64             { return globalFlagSet.Float64(name) }

func GetDuration(name string) (time.Duration, error) { return globalFlagSet.GetDuration(name) }
func Duration(name string) time.Duration             { return globalFlagSet.Duration(name) }

func GetStringSlice(name string) ([]string, error) { return globalFlagSet.GetStringSlice(name) }
func StringSlice(name string) []string             { return globalFlagSet.StringSlice(name) }

func GetIntSlice(name string) ([]int, error) { return globalFlagSet.GetIntSlice(name) }
func IntSlice(name string) []int             { return globalFlagSet.IntSlice(name) }

// --- Fluent Var Functions ---

func StringVar(p *string, name, shorthand, value, usage string) {
	globalFlagSet.StringVar(p, name, shorthand, value, usage)
}

func BoolVar(p *bool, name, shorthand string, value bool, usage string) {
	globalFlagSet.BoolVar(p, name, shorthand, value, usage)
}

func IntVar(p *int, name, shorthand string, value int, usage string) {
	globalFlagSet.IntVar(p, name, shorthand, value, usage)
}

func DurationVar(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	globalFlagSet.DurationVar(p, name, shorthand, value, usage)
}

func StringSliceVar(p *[]string, name, shorthand string, value []string, usage string) {
	globalFlagSet.StringSliceVar(p, name, shorthand, value, usage)
}
