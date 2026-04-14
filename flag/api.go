package flag

func Register(fs ...Flag) {
	flagSet.Register(fs...)
}

func Parse() error {
	return flagSet.Parse()
}

func BoolE(name string) (bool, error) { return flagSet.BoolE(name) }

// Bool parses bool flag of the flagset.
func Bool(name string) bool { return flagSet.Bool(name) }

// StringE parses string flag of the flagset with error returned.
func StringE(name string) (string, error) { return flagSet.StringE(name) }

// String parses string flag of the flagset.
func String(name string) string { return flagSet.String(name) }

// IntE parses int flag of the flagset with error returned.
func IntE(name string) (int64, error) { return flagSet.IntE(name) }

// Int parses int flag of the flagset.
func Int(name string) int64 { return flagSet.Int(name) }

// UintE parses int flag of the flagset with error returned.
func UintE(name string) (uint64, error) { return flagSet.UintE(name) }

// Uint parses int flag of the flagset.
func Uint(name string) uint64 { return flagSet.Uint(name) }

// Float64E parses int flag of the flagset with error returned.
func Float64E(name string) (float64, error) { return flagSet.Float64E(name) }

// Float64 parses int flag of the flagset.
func Float64(name string) float64 { return flagSet.Float64(name) }
