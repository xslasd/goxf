package flag

import (
	"strings"
	"time"
)

// Flag is the interface for all CLI flags that can be applied to a FlagSet.
type Flag interface {
	Apply(fs *FlagSet)
}

// parseNameAndShorthand parses the flag name and shorthand.
// If Name contains comma-separated values (e.g. "config,c" or "c,config"),
// it extracts the single-character string as the shorthand and the longer string as the name.
func parseNameAndShorthand(name, shorthand string) (string, string) {
	if strings.Contains(name, ",") {
		parts := strings.Split(name, ",")
		var mainName, short string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if len(p) == 1 && short == "" {
				short = p
			} else if mainName == "" {
				mainName = p
			}
		}
		if shorthand != "" {
			short = shorthand
		}
		return mainName, short
	}
	return strings.TrimSpace(name), strings.TrimSpace(shorthand)
}

// StringFlag represents a string CLI flag.
type StringFlag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    string
	Required   bool
	Persistent bool
	Variable   *string
	Action     func(name string, fs *FlagSet)
}

func (f *StringFlag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.StringVarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.StringVarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.StringVar(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.StringVar(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.StringP(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.StringP(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.String(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.String(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// BoolFlag represents a boolean CLI flag.
type BoolFlag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    bool
	Required   bool
	Persistent bool
	Variable   *bool
	Action     func(name string, fs *FlagSet)
}

func (f *BoolFlag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.BoolVarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.BoolVarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.BoolVar(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.BoolVar(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.BoolP(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.BoolP(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Bool(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Bool(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// IntFlag represents an integer CLI flag.
type IntFlag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    int
	Required   bool
	Persistent bool
	Variable   *int
	Action     func(name string, fs *FlagSet)
}

func (f *IntFlag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.IntVarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.IntVarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.IntVar(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.IntVar(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.IntP(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.IntP(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Int(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Int(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// Int64Flag represents an int64 CLI flag.
type Int64Flag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    int64
	Required   bool
	Persistent bool
	Variable   *int64
	Action     func(name string, fs *FlagSet)
}

func (f *Int64Flag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.Int64VarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Int64VarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Int64Var(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Int64Var(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.Int64P(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Int64P(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Int64(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Int64(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// UintFlag represents an unsigned integer CLI flag.
type UintFlag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    uint
	Required   bool
	Persistent bool
	Variable   *uint
	Action     func(name string, fs *FlagSet)
}

func (f *UintFlag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.UintVarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.UintVarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.UintVar(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.UintVar(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.UintP(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.UintP(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Uint(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Uint(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// Uint64Flag represents an unsigned uint64 CLI flag.
type Uint64Flag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    uint64
	Required   bool
	Persistent bool
	Variable   *uint64
	Action     func(name string, fs *FlagSet)
}

func (f *Uint64Flag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.Uint64VarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Uint64VarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Uint64Var(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Uint64Var(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.Uint64P(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Uint64P(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Uint64(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Uint64(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// Float64Flag represents a float64 CLI flag.
type Float64Flag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    float64
	Required   bool
	Persistent bool
	Variable   *float64
	Action     func(name string, fs *FlagSet)
}

func (f *Float64Flag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.Float64VarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Float64VarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Float64Var(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Float64Var(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.Float64P(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Float64P(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Float64(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Float64(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// DurationFlag represents a time.Duration CLI flag.
type DurationFlag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    time.Duration
	Required   bool
	Persistent bool
	Variable   *time.Duration
	Action     func(name string, fs *FlagSet)
}

func (f *DurationFlag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.DurationVarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.DurationVarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.DurationVar(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.DurationVar(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.DurationP(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.DurationP(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.Duration(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.Duration(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// StringSliceFlag represents a string slice ([]string) CLI flag.
type StringSliceFlag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    []string
	Required   bool
	Persistent bool
	Variable   *[]string
	Action     func(name string, fs *FlagSet)
}

func (f *StringSliceFlag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.StringSliceVarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.StringSliceVarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.StringSliceVar(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.StringSliceVar(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.StringSliceP(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.StringSliceP(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.StringSlice(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.StringSlice(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}

// IntSliceFlag represents an int slice ([]int) CLI flag.
type IntSliceFlag struct {
	Name       string
	Shorthand  string
	Usage      string
	EnvVar     string
	Default    []int
	Required   bool
	Persistent bool
	Variable   *[]int
	Action     func(name string, fs *FlagSet)
}

func (f *IntSliceFlag) Apply(fs *FlagSet) {
	name, short := parseNameAndShorthand(f.Name, f.Shorthand)

	if f.Variable != nil {
		if short != "" {
			fs.flags.IntSliceVarP(f.Variable, name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.IntSliceVarP(f.Variable, name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.IntSliceVar(f.Variable, name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.IntSliceVar(f.Variable, name, f.Default, f.Usage)
			}
		}
	} else {
		if short != "" {
			fs.flags.IntSliceP(name, short, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.IntSliceP(name, short, f.Default, f.Usage)
			}
		} else {
			fs.flags.IntSlice(name, f.Default, f.Usage)
			if f.Persistent {
				fs.persistentFlags.IntSlice(name, f.Default, f.Usage)
			}
		}
	}

	if f.Required {
		_ = fs.flags.SetAnnotation(name, "cobra_annotation_required", []string{"true"})
	}

	fs.registerMetadata(name, f.EnvVar, f.Action)
}
