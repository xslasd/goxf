package flag

import (
	"context"
	"os"
	"strings"
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
	for _, cmd := range cmds {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		globalRootCmd.AddCommand(cmd)
	}
}

// AddCommandWithBase 注册一个或多个子命令到全局根命令，并在执行前自动完成 InitBase (配置加载、运行时元数据、日志系统)
func AddCommandWithBase(cmds ...*cobra.Command) {
	globalMu.Lock()
	defer globalMu.Unlock()
	for _, cmd := range cmds {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		RequireInitBase(cmd, true)
		globalRootCmd.AddCommand(cmd)
	}
}

// AddBaseCommand 是 AddCommandWithBase 的简短别名
func AddBaseCommand(cmds ...*cobra.Command) {
	AddCommandWithBase(cmds...)
}

// AddCommandByRequireInitBase 是 AddCommandWithBase 的兼容别名
func AddCommandByRequireInitBase(cmds ...*cobra.Command) {
	AddCommandWithBase(cmds...)
}

// AddCommandWithBaseExcept 注册子命令并在执行前自动完成 InitBase，但若命令行中输入了指定的排除参数（如 gen / g），则自动跳过 InitBase
func AddCommandWithBaseExcept(cmd *cobra.Command, exceptFlags ...string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	RequireInitBase(cmd, true)
	if len(exceptFlags) > 0 {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[AnnotationExceptFlags] = strings.Join(exceptFlags, ",")
	}
	globalRootCmd.AddCommand(cmd)
}

// AddBaseCommandExcept 是 AddCommandWithBaseExcept 的简短别名
func AddBaseCommandExcept(cmd *cobra.Command, exceptFlags ...string) {
	AddCommandWithBaseExcept(cmd, exceptFlags...)
}

// AddWrappedSubCommand adds one or more wrapped *Command items directly to the global root command, automatically preparing flags.
func AddWrappedSubCommand(cmds ...*Command) {
	globalMu.Lock()
	defer globalMu.Unlock()
	for _, cmd := range cmds {
		cmd.Command.SilenceErrors = true
		cmd.Command.SilenceUsage = true
		cmd.prepare()
		globalRootCmd.AddCommand(cmd.Command)
	}
}

// AddWrappedCommand is an alias for AddWrappedSubCommand.
func AddWrappedCommand(cmds ...*Command) {
	AddWrappedSubCommand(cmds...)
}

// AddWrappedCommandWithBase 注册包装子命令并在执行前自动完成 InitBase
func AddWrappedCommandWithBase(cmds ...*Command) {
	globalMu.Lock()
	defer globalMu.Unlock()
	for _, cmd := range cmds {
		cmd.Command.SilenceErrors = true
		cmd.Command.SilenceUsage = true
		cmd.RequireInitBase(true)
		cmd.prepare()
		globalRootCmd.AddCommand(cmd.Command)
	}
}

// AddWrappedBaseCommand 是 AddWrappedCommandWithBase 的简短别名
func AddWrappedBaseCommand(cmds ...*Command) {
	AddWrappedCommandWithBase(cmds...)
}

// AddWrappedCommandByRequireInitBase 是 AddWrappedCommandWithBase 的兼容别名
func AddWrappedCommandByRequireInitBase(cmds ...*Command) {
	AddWrappedCommandWithBase(cmds...)
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

const (
	// AnnotationRequireInitBase is the Cobra Command annotation key used to indicate whether InitBase is required before execution.
	AnnotationRequireInitBase = "goxf:require_init_base"
	// AnnotationExceptFlags is the Cobra Command annotation key used to specify flags that skip InitBase when present in CLI invocation.
	AnnotationExceptFlags = "goxf:init_base_except_flags"
)

// RequireInitBase marks a cobra.Command as requiring goxf InitBase (config, runtime, logger) before execution.
func RequireInitBase(cmd *cobra.Command, require ...bool) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	val := "true"
	if len(require) > 0 && !require[0] {
		val = "false"
	}
	cmd.Annotations[AnnotationRequireInitBase] = val
	return cmd
}

// ShouldInitBase checks if the given command explicitly requests InitBase before execution.
func ShouldInitBase(cmd *cobra.Command) bool {
	if cmd == nil || cmd.Annotations == nil {
		return false
	}
	if cmd.Annotations[AnnotationRequireInitBase] != "true" {
		return false
	}
	// 检查是否有排除标志
	if except := cmd.Annotations[AnnotationExceptFlags]; except != "" {
		for _, flagName := range strings.Split(except, ",") {
			flagName = strings.TrimSpace(flagName)
			if flagName == "" {
				continue
			}
			for _, arg := range os.Args[1:] {
				if arg == "-"+flagName || arg == "--"+flagName ||
					strings.HasPrefix(arg, "-"+flagName+"=") ||
					strings.HasPrefix(arg, "--"+flagName+"=") {
					return false // 命中排除参数，跳过 InitBase！
				}
			}
		}
	}
	return true
}

// MatchedSubCommand returns the matched subcommand from CLI args, or nil if none matched or root command matched.
func MatchedSubCommand() *cobra.Command {
	if len(os.Args) <= 1 {
		return nil
	}
	globalMu.RLock()
	defer globalMu.RUnlock()
	cmd, _, err := globalRootCmd.Find(os.Args[1:])
	if err != nil || cmd == nil || cmd == globalRootCmd {
		return nil
	}
	return cmd
}

// HasSubCommand checks if the current CLI invocation matches any registered subcommands (excluding root).
func HasSubCommand() bool {
	return MatchedSubCommand() != nil
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
