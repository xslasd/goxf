package flag

import (
	"context"

	"github.com/spf13/cobra"
)

// Command represents a CLI command, wrapping cobra.Command.
type Command struct {
	*cobra.Command
	flagSet *FlagSet
}

// NewCommand creates a new Command with a given use and short description.
func NewCommand(use, short string) *Command {
	c := &cobra.Command{
		Use:   use,
		Short: short,
	}
	cmd := &Command{
		Command: c,
		flagSet: NewFlagSet(use),
	}
	return cmd
}

// WrapCommand wraps an existing cobra.Command into a goxf Command.
func WrapCommand(c *cobra.Command) *Command {
	return &Command{
		Command: c,
		flagSet: NewFlagSet(c.Use),
	}
}

// Flags returns the local FlagSet for this command.
func (c *Command) FlagSet() *FlagSet {
	return c.flagSet
}

// Register registers flags to this command.
func (c *Command) Register(flags ...Flag) *Command {
	c.flagSet.Register(flags...)
	return c
}

// BindStruct binds struct fields with tags to this command's flags.
func (c *Command) BindStruct(v interface{}) *Command {
	_ = c.flagSet.BindStruct(v)
	return c
}

// RequireInitBase marks whether this command requires goxf InitBase (config, runtime, logger) before execution.
func (c *Command) RequireInitBase(require ...bool) *Command {
	RequireInitBase(c.Command, require...)
	return c
}

// AddSubCommand adds subcommands to this command.
func (c *Command) AddSubCommand(cmds ...*cobra.Command) *Command {
	c.Command.AddCommand(cmds...)
	return c
}

// AddWrappedSubCommand adds wrapped commands to this command.
func (c *Command) AddWrappedSubCommand(cmds ...*Command) *Command {
	for _, cmd := range cmds {
		cmd.prepare()
		c.Command.AddCommand(cmd.Command)
	}
	return c
}

// prepare syncs registered flags from our FlagSet to cobra.Command's Flags before execution.
func (c *Command) prepare() {
	c.flagSet.applyAll()
	// Merge local flags
	c.Command.Flags().AddFlagSet(c.flagSet.Flags())
	// Merge persistent flags
	c.Command.PersistentFlags().AddFlagSet(c.flagSet.PersistentFlags())
}

// Execute prepares flags and runs the cobra command.
func (c *Command) Execute() error {
	c.prepare()
	return c.Command.Execute()
}

// ExecuteContext prepares flags and runs the cobra command with context.
func (c *Command) ExecuteContext(ctx context.Context) error {
	c.prepare()
	return c.Command.ExecuteContext(ctx)
}
