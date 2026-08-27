package flag

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
)

// FlagSet wraps pflag.FlagSet with environment variables, custom actions,
// and struct-tag binding capabilities.
type FlagSet struct {
	mu              sync.RWMutex
	name            string
	flags           *pflag.FlagSet
	persistentFlags *pflag.FlagSet
	registeredFlags []Flag
	actions         map[string]func(string, *FlagSet)
	environs        map[string]string // flagName -> envVar
	parsed          bool
}

// NewFlagSet creates a new FlagSet.
func NewFlagSet(name string) *FlagSet {
	return &FlagSet{
		name:            name,
		flags:           pflag.NewFlagSet(name, pflag.ContinueOnError),
		persistentFlags: pflag.NewFlagSet(name+"-persistent", pflag.ContinueOnError),
		registeredFlags: make([]Flag, 0),
		actions:         make(map[string]func(string, *FlagSet)),
		environs:        make(map[string]string),
	}
}

// Name returns the name of the FlagSet.
func (fs *FlagSet) Name() string {
	return fs.name
}

// Flags returns the underlying pflag.FlagSet.
func (fs *FlagSet) Flags() *pflag.FlagSet {
	return fs.flags
}

// PersistentFlags returns the underlying pflag.FlagSet for persistent flags.
func (fs *FlagSet) PersistentFlags() *pflag.FlagSet {
	return fs.persistentFlags
}

// registerMetadata registers environment variable mapping and action for a flag.
func (fs *FlagSet) registerMetadata(name, envVar string, action func(string, *FlagSet)) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if envVar != "" {
		fs.environs[name] = envVar
	}
	if action != nil {
		fs.actions[name] = action
	}
}

// Register registers one or more Flag implementations.
func (fs *FlagSet) Register(flags ...Flag) {
	fs.mu.Lock()
	fs.registeredFlags = append(fs.registeredFlags, flags...)
	fs.mu.Unlock()
}

// applyAll applies all registered Flag objects to underlying pflags.
func (fs *FlagSet) applyAll() {
	fs.mu.RLock()
	flags := make([]Flag, len(fs.registeredFlags))
	copy(flags, fs.registeredFlags)
	fs.mu.RUnlock()

	for _, f := range flags {
		f.Apply(fs)
	}
}

// Parse parses flag definitions from the argument list (defaults to os.Args[1:] if empty).
func (fs *FlagSet) Parse(args ...string) error {
	if len(args) == 0 {
		return fs.ParseArgs(os.Args[1:])
	}
	return fs.ParseArgs(args)
}

// ParseArgs parses flag definitions from the specified argument list.
func (fs *FlagSet) ParseArgs(args []string) error {
	fs.applyAll()

	// Parse flags using main FlagSet which contains all registered flags
	if err := fs.flags.Parse(args); err != nil {
		return err
	}

	fs.mu.Lock()
	fs.parsed = true

	// Apply environment variable fallback if flag was not explicitly changed on CLI
	for flagName, envVar := range fs.environs {
		f := fs.flags.Lookup(flagName)
		if f != nil && !f.Changed {
			if envVal, ok := os.LookupEnv(envVar); ok && envVal != "" {
				_ = f.Value.Set(envVal)
			}
		}
	}

	// Trigger actions for flags that were explicitly changed on CLI
	actions := make(map[string]func(string, *FlagSet), len(fs.actions))
	for k, v := range fs.actions {
		actions[k] = v
	}
	fs.mu.Unlock()

	for flagName, action := range actions {
		f := fs.Lookup(flagName)
		if f != nil && f.Changed {
			action(flagName, fs)
		}
	}

	return nil
}

// Parsed returns whether the FlagSet has been parsed.
func (fs *FlagSet) Parsed() bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.parsed
}

// Lookup searches for a flag by name.
func (fs *FlagSet) Lookup(name string) *pflag.Flag {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.flags.Lookup(name)
}

// Changed returns true if the flag was explicitly set on the command line.
func (fs *FlagSet) Changed(name string) bool {
	f := fs.Lookup(name)
	return f != nil && f.Changed
}

// Has returns true if the flag exists in the FlagSet.
func (fs *FlagSet) Has(name string) bool {
	return fs.Lookup(name) != nil
}

// PrintDefaults prints default values of all defined flags.
func (fs *FlagSet) PrintDefaults() {
	fs.flags.PrintDefaults()
}

// VisitAll visits all flags in alphabetical order.
func (fs *FlagSet) VisitAll(fn func(*pflag.Flag)) {
	fs.flags.VisitAll(fn)
}

// --- Getter Methods with Error and Safe Helpers ---

// GetString returns the string value of a flag.
func (fs *FlagSet) GetString(name string) (string, error) {
	f := fs.Lookup(name)
	if f == nil {
		return "", fmt.Errorf("flag accessed but not defined: %s", name)
	}
	return f.Value.String(), nil
}

// String returns the string value of a flag (returns empty string if error).
func (fs *FlagSet) String(name string) string {
	val, _ := fs.GetString(name)
	return val
}

// GetBool returns the bool value of a flag.
func (fs *FlagSet) GetBool(name string) (bool, error) {
	f := fs.Lookup(name)
	if f == nil {
		return false, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	return strconv.ParseBool(f.Value.String())
}

// Bool returns the bool value of a flag (returns false if error).
func (fs *FlagSet) Bool(name string) bool {
	val, _ := fs.GetBool(name)
	return val
}

// GetInt returns the int value of a flag.
func (fs *FlagSet) GetInt(name string) (int, error) {
	f := fs.Lookup(name)
	if f == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	val, err := strconv.ParseInt(f.Value.String(), 10, strconv.IntSize)
	return int(val), err
}

// Int returns the int value of a flag (returns 0 if error).
func (fs *FlagSet) Int(name string) int {
	val, _ := fs.GetInt(name)
	return val
}

// GetInt64 returns the int64 value of a flag.
func (fs *FlagSet) GetInt64(name string) (int64, error) {
	f := fs.Lookup(name)
	if f == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	return strconv.ParseInt(f.Value.String(), 10, 64)
}

// Int64 returns the int64 value of a flag (returns 0 if error).
func (fs *FlagSet) Int64(name string) int64 {
	val, _ := fs.GetInt64(name)
	return val
}

// GetUint returns the uint value of a flag.
func (fs *FlagSet) GetUint(name string) (uint, error) {
	f := fs.Lookup(name)
	if f == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	val, err := strconv.ParseUint(f.Value.String(), 10, strconv.IntSize)
	return uint(val), err
}

// Uint returns the uint value of a flag (returns 0 if error).
func (fs *FlagSet) Uint(name string) uint {
	val, _ := fs.GetUint(name)
	return val
}

// GetUint64 returns the uint64 value of a flag.
func (fs *FlagSet) GetUint64(name string) (uint64, error) {
	f := fs.Lookup(name)
	if f == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	return strconv.ParseUint(f.Value.String(), 10, 64)
}

// Uint64 returns the uint64 value of a flag (returns 0 if error).
func (fs *FlagSet) Uint64(name string) uint64 {
	val, _ := fs.GetUint64(name)
	return val
}

// GetFloat64 returns the float64 value of a flag.
func (fs *FlagSet) GetFloat64(name string) (float64, error) {
	f := fs.Lookup(name)
	if f == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	return strconv.ParseFloat(f.Value.String(), 64)
}

// Float64 returns the float64 value of a flag (returns 0 if error).
func (fs *FlagSet) Float64(name string) float64 {
	val, _ := fs.GetFloat64(name)
	return val
}

// GetDuration returns the time.Duration value of a flag.
func (fs *FlagSet) GetDuration(name string) (time.Duration, error) {
	f := fs.Lookup(name)
	if f == nil {
		return 0, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	return time.ParseDuration(f.Value.String())
}

// Duration returns the time.Duration value of a flag (returns 0 if error).
func (fs *FlagSet) Duration(name string) time.Duration {
	val, _ := fs.GetDuration(name)
	return val
}

// GetStringSlice returns the []string value of a flag.
func (fs *FlagSet) GetStringSlice(name string) ([]string, error) {
	return fs.flags.GetStringSlice(name)
}

// StringSlice returns the []string value of a flag.
func (fs *FlagSet) StringSlice(name string) []string {
	val, _ := fs.GetStringSlice(name)
	return val
}

// GetIntSlice returns the []int value of a flag.
func (fs *FlagSet) GetIntSlice(name string) ([]int, error) {
	return fs.flags.GetIntSlice(name)
}

// IntSlice returns the []int value of a flag.
func (fs *FlagSet) IntSlice(name string) []int {
	val, _ := fs.GetIntSlice(name)
	return val
}

// --- Fluent Inline Variable Registration ---

// StringVar binds a string flag with an optional shorthand.
func (fs *FlagSet) StringVar(p *string, name, shorthand, value, usage string) {
	if shorthand != "" {
		fs.flags.StringVarP(p, name, shorthand, value, usage)
	} else {
		fs.flags.StringVar(p, name, value, usage)
	}
}

// BoolVar binds a bool flag with an optional shorthand.
func (fs *FlagSet) BoolVar(p *bool, name, shorthand string, value bool, usage string) {
	if shorthand != "" {
		fs.flags.BoolVarP(p, name, shorthand, value, usage)
	} else {
		fs.flags.BoolVar(p, name, value, usage)
	}
}

// IntVar binds an int flag with an optional shorthand.
func (fs *FlagSet) IntVar(p *int, name, shorthand string, value int, usage string) {
	if shorthand != "" {
		fs.flags.IntVarP(p, name, shorthand, value, usage)
	} else {
		fs.flags.IntVar(p, name, value, usage)
	}
}

// DurationVar binds a time.Duration flag with an optional shorthand.
func (fs *FlagSet) DurationVar(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	if shorthand != "" {
		fs.flags.DurationVarP(p, name, shorthand, value, usage)
	} else {
		fs.flags.DurationVar(p, name, value, usage)
	}
}

// StringSliceVar binds a []string flag with an optional shorthand.
func (fs *FlagSet) StringSliceVar(p *[]string, name, shorthand string, value []string, usage string) {
	if shorthand != "" {
		fs.flags.StringSliceVarP(p, name, shorthand, value, usage)
	} else {
		fs.flags.StringSliceVar(p, name, value, usage)
	}
}

// --- Struct Tag Auto Binding ---

// BindStruct automatically binds struct fields to CLI flags based on struct tags.
func (fs *FlagSet) BindStruct(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("BindStruct requires a pointer to a struct, got %T", v)
	}

	elem := val.Elem()
	typ := elem.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := elem.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		flagTag := field.Tag.Get("flag")
		if flagTag == "" || flagTag == "-" {
			continue
		}

		parts := strings.Split(flagTag, ",")
		flagName := strings.TrimSpace(parts[0])
		shorthand := ""
		if len(parts) > 1 {
			shorthand = strings.TrimSpace(parts[1])
		}

		usage := field.Tag.Get("usage")
		defaultVal := field.Tag.Get("default")
		envVar := field.Tag.Get("env")
		required := field.Tag.Get("required") == "true"
		persistent := field.Tag.Get("persistent") == "true"

		switch field.Type.Kind() {
		case reflect.String:
			ptr := fieldVal.Addr().Interface().(*string)
			fs.Register(&StringFlag{
				Name:       flagName,
				Shorthand:  shorthand,
				Usage:      usage,
				Default:    defaultVal,
				EnvVar:     envVar,
				Required:   required,
				Persistent: persistent,
				Variable:   ptr,
			})
		case reflect.Bool:
			ptr := fieldVal.Addr().Interface().(*bool)
			defBool, _ := strconv.ParseBool(defaultVal)
			fs.Register(&BoolFlag{
				Name:       flagName,
				Shorthand:  shorthand,
				Usage:      usage,
				Default:    defBool,
				EnvVar:     envVar,
				Required:   required,
				Persistent: persistent,
				Variable:   ptr,
			})
		case reflect.Int:
			ptr := fieldVal.Addr().Interface().(*int)
			defInt, _ := strconv.Atoi(defaultVal)
			fs.Register(&IntFlag{
				Name:       flagName,
				Shorthand:  shorthand,
				Usage:      usage,
				Default:    defInt,
				EnvVar:     envVar,
				Required:   required,
				Persistent: persistent,
				Variable:   ptr,
			})
		case reflect.Int64:
			if field.Type == reflect.TypeOf(time.Duration(0)) {
				ptr := fieldVal.Addr().Interface().(*time.Duration)
				defDur, _ := time.ParseDuration(defaultVal)
				fs.Register(&DurationFlag{
					Name:       flagName,
					Shorthand:  shorthand,
					Usage:      usage,
					Default:    defDur,
					EnvVar:     envVar,
					Required:   required,
					Persistent: persistent,
					Variable:   ptr,
				})
			} else {
				ptr := fieldVal.Addr().Interface().(*int64)
				defInt64, _ := strconv.ParseInt(defaultVal, 10, 64)
				fs.Register(&Int64Flag{
					Name:       flagName,
					Shorthand:  shorthand,
					Usage:      usage,
					Default:    defInt64,
					EnvVar:     envVar,
					Required:   required,
					Persistent: persistent,
					Variable:   ptr,
				})
			}
		case reflect.Uint:
			ptr := fieldVal.Addr().Interface().(*uint)
			defUint64, _ := strconv.ParseUint(defaultVal, 10, strconv.IntSize)
			fs.Register(&UintFlag{
				Name:       flagName,
				Shorthand:  shorthand,
				Usage:      usage,
				Default:    uint(defUint64),
				EnvVar:     envVar,
				Required:   required,
				Persistent: persistent,
				Variable:   ptr,
			})
		case reflect.Float64:
			ptr := fieldVal.Addr().Interface().(*float64)
			defFloat, _ := strconv.ParseFloat(defaultVal, 64)
			fs.Register(&Float64Flag{
				Name:       flagName,
				Shorthand:  shorthand,
				Usage:      usage,
				Default:    defFloat,
				EnvVar:     envVar,
				Required:   required,
				Persistent: persistent,
				Variable:   ptr,
			})
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				ptr := fieldVal.Addr().Interface().(*[]string)
				var defSlice []string
				if defaultVal != "" {
					defSlice = strings.Split(defaultVal, ",")
				}
				fs.Register(&StringSliceFlag{
					Name:       flagName,
					Shorthand:  shorthand,
					Usage:      usage,
					Default:    defSlice,
					EnvVar:     envVar,
					Required:   required,
					Persistent: persistent,
					Variable:   ptr,
				})
			}
		}
	}
	return nil
}
