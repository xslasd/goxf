package flag

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type FlagSet struct {
	*flag.FlagSet
	flags    []Flag
	actions  map[string]func(string, *FlagSet)
	environs map[string]string
}

var flagSet = &FlagSet{
	FlagSet:  flag.CommandLine,
	flags:    make([]Flag, 0),
	actions:  make(map[string]func(string, *FlagSet)),
	environs: make(map[string]string),
}

// Register ...
func (fs *FlagSet) Register(flags ...Flag) {
	fs.flags = append(fs.flags, flags...)
}

func (fs *FlagSet) Parse() error {
	if fs.Parsed() {
		return nil
	}
	for _, f := range fs.flags {
		f.Apply(fs)
	}

	if err := fs.FlagSet.Parse(os.Args[1:]); err != nil {
		return err
	}

	fs.FlagSet.Visit(func(f *flag.Flag) {
		if action, ok := fs.actions[f.Name]; ok && action != nil {
			action(f.Name, fs)
		}
	})

	return nil
}

func (fs *FlagSet) Lookup(name string) *flag.Flag {
	flagI := fs.FlagSet.Lookup(name)
	if flagI != nil {
		if flagI.Value.String() == "" {
			if env, ok := fs.environs[name]; ok {
				_ = flagI.Value.Set(env)
			}
		}
		if flagI.Value.String() == "" {
			_ = flagI.Value.Set(flagI.DefValue)
		}
	}
	return flagI
}

// BoolE parses bool flag of provided flagset with error returned.
func (fs *FlagSet) BoolE(name string) (bool, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseBool(flag.Value.String())
	}

	return false, fmt.Errorf("undefined flag name: %s", name)
}

// Bool parses bool flag of provided flagset.
func (fs *FlagSet) Bool(name string) bool {
	ret, _ := fs.BoolE(name)
	return ret
}

// StringE parses string flag of provided flagset with error returned.
func (fs *FlagSet) StringE(name string) (string, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return flag.Value.String(), nil
	}

	return "", fmt.Errorf("undefined flag name: %s", name)
}

// String parses string flag of provided flagset.
func (fs *FlagSet) String(name string) string {
	ret, _ := fs.StringE(name)
	return ret
}

// IntE parses int flag of provided flagset with error returned.
func (fs *FlagSet) IntE(name string) (int64, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseInt(flag.Value.String(), 10, 64)
	}

	return 0, fmt.Errorf("undefined flag name: %s", name)
}

// Int parses int flag of provided flagset.
func (fs *FlagSet) Int(name string) int64 {
	ret, _ := fs.IntE(name)
	return ret
}

// UintE parses int flag of provided flagset with error returned.
func (fs *FlagSet) UintE(name string) (uint64, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseUint(flag.Value.String(), 10, 64)
	}

	return 0, fmt.Errorf("undefined flag name: %s", name)
}

// Uint parses int flag of provided flagset.
func (fs *FlagSet) Uint(name string) uint64 {
	ret, _ := fs.UintE(name)
	return ret
}

// Float64E parses int flag of provided flagset with error returned.
func (fs *FlagSet) Float64E(name string) (float64, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseFloat(flag.Value.String(), 64)
	}

	return 0.0, fmt.Errorf("undefined flag name: %s", name)
}

// Float64 parses int flag of provided flagset.
func (fs *FlagSet) Float64(name string) float64 {
	ret, _ := fs.Float64E(name)
	return ret
}
