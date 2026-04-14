package flag

import (
	"os"
	"strings"
)

type Flag interface {
	Apply(*FlagSet)
}

// BoolFlag is a bool flag implements of Flag interface.
type BoolFlag struct {
	Name     string
	Usage    string
	EnvVar   string
	Default  bool
	Variable *bool
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *BoolFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.BoolVar(f.Variable, field, f.Default, f.Usage)
		}

		set.FlagSet.Bool(field, f.Default, f.Usage)
		set.actions[field] = f.Action
		set.environs[field] = os.Getenv(f.EnvVar)
	}
}

// StringFlag is a string flag implements of Flag interface.
type StringFlag struct {
	Name     string
	Usage    string
	EnvVar   string
	Default  string
	Variable *string
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *StringFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.StringVar(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.String(field, f.Default, f.Usage)
		set.actions[field] = f.Action
		set.environs[field] = os.Getenv(f.EnvVar)
	}
}

// IntFlag is an int flag implements of Flag interface.
type IntFlag struct {
	Name     string
	Usage    string
	Default  int
	Variable *int
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *IntFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.IntVar(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.Int(field, f.Default, f.Usage)
		set.actions[field] = f.Action
	}
}

// UintFlag is an uint flag implements of Flag interface.
type UintFlag struct {
	Name     string
	Usage    string
	Default  uint
	Variable *uint
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *UintFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.UintVar(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.Uint(field, f.Default, f.Usage)
		set.actions[field] = f.Action
	}
}

// Float64Flag is a float flag implements of Flag interface.
type Float64Flag struct {
	Name     string
	Usage    string
	Default  float64
	Variable *float64
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *Float64Flag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.Float64Var(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.Float64(field, f.Default, f.Usage)
		set.actions[field] = f.Action
	}
}
