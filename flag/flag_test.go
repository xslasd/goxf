package flag

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlag_BasicTypesAndDefaults(t *testing.T) {
	fs := NewFlagSet("test")

	var (
		strVal      string
		boolVal     bool
		intVal      int
		int64Val    int64
		uintVal     uint
		uint64Val   uint64
		floatVal    float64
		durVal      time.Duration
		strSliceVal []string
		intSliceVal []int
	)

	fs.Register(
		&StringFlag{Name: "config", Shorthand: "c", Default: "app.yaml", Variable: &strVal},
		&BoolFlag{Name: "watch", Shorthand: "w", Default: true, Variable: &boolVal},
		&IntFlag{Name: "port", Shorthand: "p", Default: 8080, Variable: &intVal},
		&Int64Flag{Name: "count", Default: int64(100000), Variable: &int64Val},
		&UintFlag{Name: "retry", Default: uint(3), Variable: &uintVal},
		&Uint64Flag{Name: "max-size", Default: uint64(1048576), Variable: &uint64Val},
		&Float64Flag{Name: "ratio", Default: 0.75, Variable: &floatVal},
		&DurationFlag{Name: "timeout", Shorthand: "t", Default: 5 * time.Second, Variable: &durVal},
		&StringSliceFlag{Name: "tags", Default: []string{"api", "v1"}, Variable: &strSliceVal},
		&IntSliceFlag{Name: "codes", Default: []int{200, 201}, Variable: &intSliceVal},
	)

	err := fs.ParseArgs([]string{})
	require.NoError(t, err)

	// Check variable bindings with defaults
	assert.Equal(t, "app.yaml", strVal)
	assert.True(t, boolVal)
	assert.Equal(t, 8080, intVal)
	assert.Equal(t, int64(100000), int64Val)
	assert.Equal(t, uint(3), uintVal)
	assert.Equal(t, uint64(1048576), uint64Val)
	assert.Equal(t, 0.75, floatVal)
	assert.Equal(t, 5*time.Second, durVal)
	assert.Equal(t, []string{"api", "v1"}, strSliceVal)
	assert.Equal(t, []int{200, 201}, intSliceVal)

	// Check getter methods
	assert.Equal(t, "app.yaml", fs.String("config"))
	assert.True(t, fs.Bool("watch"))
	assert.Equal(t, 8080, fs.Int("port"))
	assert.Equal(t, int64(100000), fs.Int64("count"))
	assert.Equal(t, uint(3), fs.Uint("retry"))
	assert.Equal(t, uint64(1048576), fs.Uint64("max-size"))
	assert.Equal(t, 0.75, fs.Float64("ratio"))
	assert.Equal(t, 5*time.Second, fs.Duration("timeout"))
	assert.Equal(t, []string{"api", "v1"}, fs.StringSlice("tags"))
	assert.Equal(t, []int{200, 201}, fs.IntSlice("codes"))

	// Check Get* with error helpers
	sVal, err := fs.GetString("config")
	assert.NoError(t, err)
	assert.Equal(t, "app.yaml", sVal)

	bVal, err := fs.GetBool("watch")
	assert.NoError(t, err)
	assert.True(t, bVal)

	iVal, err := fs.GetInt("port")
	assert.NoError(t, err)
	assert.Equal(t, 8080, iVal)

	i64Val, err := fs.GetInt64("count")
	assert.NoError(t, err)
	assert.Equal(t, int64(100000), i64Val)

	uVal, err := fs.GetUint("retry")
	assert.NoError(t, err)
	assert.Equal(t, uint(3), uVal)

	u64Val, err := fs.GetUint64("max-size")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1048576), u64Val)

	fVal, err := fs.GetFloat64("ratio")
	assert.NoError(t, err)
	assert.Equal(t, 0.75, fVal)

	dVal, err := fs.GetDuration("timeout")
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, dVal)

	ssVal, err := fs.GetStringSlice("tags")
	assert.NoError(t, err)
	assert.Equal(t, []string{"api", "v1"}, ssVal)

	isVal, err := fs.GetIntSlice("codes")
	assert.NoError(t, err)
	assert.Equal(t, []int{200, 201}, isVal)

	assert.Equal(t, "test", fs.Name())
	assert.NotNil(t, fs.Flags())
	assert.NotNil(t, fs.PersistentFlags())
	assert.True(t, fs.Parsed())
}

func TestFlag_CLIOverrideAndShorthand(t *testing.T) {
	fs := NewFlagSet("test")

	fs.Register(
		&StringFlag{Name: "config,c", Default: "default.yaml"},
		&BoolFlag{Name: "watch,w", Default: false},
		&IntFlag{Name: "port,p", Default: 8080},
		&Int64Flag{Name: "max-rows,m", Default: 100},
		&UintFlag{Name: "count,u", Default: 1},
		&Uint64Flag{Name: "bytes,b", Default: 1024},
		&Float64Flag{Name: "rate,r", Default: 1.0},
		&DurationFlag{Name: "timeout,t", Default: 1 * time.Second},
		&StringSliceFlag{Name: "nodes,n", Default: []string{"node1"}},
		&IntSliceFlag{Name: "ids,i", Default: []int{1}},
	)

	err := fs.ParseArgs([]string{
		"-c", "custom.yaml",
		"-w",
		"-p", "9090",
		"-m", "5000",
		"-u", "10",
		"-b", "2048",
		"-r", "2.5",
		"-t", "10s",
		"-n", "node2,node3",
		"-i", "2,3,4",
	})
	require.NoError(t, err)

	assert.Equal(t, "custom.yaml", fs.String("config"))
	assert.True(t, fs.Bool("watch"))
	assert.Equal(t, 9090, fs.Int("port"))
	assert.Equal(t, int64(5000), fs.Int64("max-rows"))
	assert.Equal(t, uint(10), fs.Uint("count"))
	assert.Equal(t, uint64(2048), fs.Uint64("bytes"))
	assert.Equal(t, 2.5, fs.Float64("rate"))
	assert.Equal(t, 10*time.Second, fs.Duration("timeout"))
	assert.Equal(t, []string{"node2", "node3"}, fs.StringSlice("nodes"))
	assert.Equal(t, []int{2, 3, 4}, fs.IntSlice("ids"))

	assert.True(t, fs.Changed("config"))
	assert.True(t, fs.Changed("watch"))
}

func TestFlag_EnvironmentPrecedence(t *testing.T) {
	const envKey = "GOXF_TEST_ENV_CONFIG"
	_ = os.Setenv(envKey, "env_config.yaml")
	defer func() { _ = os.Unsetenv(envKey) }()

	// Case 1: Environment variable overrides default
	fs1 := NewFlagSet("test1")
	fs1.Register(&StringFlag{
		Name:    "config",
		Default: "default.yaml",
		EnvVar:  envKey,
	})
	err := fs1.ParseArgs([]string{})
	require.NoError(t, err)
	assert.Equal(t, "env_config.yaml", fs1.String("config"))

	// Case 2: CLI argument overrides environment variable
	fs2 := NewFlagSet("test2")
	fs2.Register(&StringFlag{
		Name:    "config",
		Default: "default.yaml",
		EnvVar:  envKey,
	})
	err = fs2.ParseArgs([]string{"--config", "cli_override.yaml"})
	require.NoError(t, err)
	assert.Equal(t, "cli_override.yaml", fs2.String("config"))
}

func TestFlag_ActionCallback(t *testing.T) {
	fs := NewFlagSet("test")

	actionTriggered := false
	var triggeredFlagName string

	fs.Register(&BoolFlag{
		Name:    "version",
		Default: false,
		Action: func(name string, f *FlagSet) {
			actionTriggered = true
			triggeredFlagName = name
		},
	})

	err := fs.ParseArgs([]string{"--version"})
	require.NoError(t, err)
	assert.True(t, actionTriggered)
	assert.Equal(t, "version", triggeredFlagName)
}

func TestFlag_BindStruct(t *testing.T) {
	type ServerConfig struct {
		ConfigPath string        `flag:"config,c" default:"conf/app.yaml" usage:"path to config"`
		Debug      bool          `flag:"debug,d" default:"false" usage:"enable debug mode"`
		Port       int           `flag:"port,p" default:"8000" usage:"server port"`
		Count      int64         `flag:"count" default:"500" usage:"item count"`
		Workers    uint          `flag:"workers" default:"4" usage:"worker count"`
		Rate       float64       `flag:"rate" default:"1.5" usage:"sample rate"`
		Timeout    time.Duration `flag:"timeout,t" default:"3s" usage:"request timeout"`
		Hosts      []string      `flag:"hosts" default:"127.0.0.1,localhost" usage:"allowed hosts"`
	}

	cfg := &ServerConfig{}
	fs := NewFlagSet("bind_test")
	err := fs.BindStruct(cfg)
	require.NoError(t, err)

	err = fs.ParseArgs([]string{
		"-c", "production.yaml",
		"-d",
		"-p", "9999",
		"--count", "1000",
		"--workers", "8",
		"--rate", "3.14",
		"-t", "15s",
	})
	require.NoError(t, err)

	assert.Equal(t, "production.yaml", cfg.ConfigPath)
	assert.True(t, cfg.Debug)
	assert.Equal(t, 9999, cfg.Port)
	assert.Equal(t, int64(1000), cfg.Count)
	assert.Equal(t, uint(8), cfg.Workers)
	assert.Equal(t, 3.14, cfg.Rate)
	assert.Equal(t, 15*time.Second, cfg.Timeout)
	assert.Equal(t, []string{"127.0.0.1", "localhost"}, cfg.Hosts)
}

func TestCommand_SubCommands(t *testing.T) {
	rootCmd := NewCommand("app", "root application")

	var runCalled bool
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run service",
		Run: func(cmd *cobra.Command, args []string) {
			runCalled = true
		},
	}

	var cfgPath string
	rootCmd.Register(&StringFlag{
		Name:       "config",
		Shorthand:  "c",
		Default:    "app.yaml",
		Persistent: true,
		Variable:   &cfgPath,
	})

	rootCmd.AddSubCommand(runCmd)
	rootCmd.SetArgs([]string{"run", "-c", "custom.yaml"})

	err := rootCmd.Execute()
	require.NoError(t, err)
	assert.True(t, runCalled)
	assert.Equal(t, "custom.yaml", cfgPath)
	assert.NotNil(t, rootCmd.FlagSet())
}

func TestCommand_WrapCommandAndWrappedSubCommand(t *testing.T) {
	rawCmd := &cobra.Command{Use: "service"}
	wrappedRoot := WrapCommand(rawCmd)

	subWrapped := NewCommand("worker", "worker runner")
	var executed bool
	subWrapped.Run = func(cmd *cobra.Command, args []string) {
		executed = true
	}

	wrappedRoot.AddWrappedSubCommand(subWrapped)
	wrappedRoot.SetArgs([]string{"worker"})

	err := wrappedRoot.ExecuteContext(context.Background())
	require.NoError(t, err)
	assert.True(t, executed)
}

func TestGlobalAPI(t *testing.T) {
	Reset()

	var (
		cfg   string
		watch bool
		port  int
		dur   time.Duration
		nodes []string
	)

	StringVar(&cfg, "config", "c", "init.yaml", "config file")
	BoolVar(&watch, "watch", "w", false, "watch mode")
	IntVar(&port, "port", "p", 8080, "server port")
	DurationVar(&dur, "timeout", "t", 5*time.Second, "timeout")
	StringSliceVar(&nodes, "nodes", "n", []string{"n1"}, "nodes")

	Register(&BoolFlag{Name: "verbose,v", Default: false})

	err := ParseArgs([]string{"-c", "global.yaml", "-w", "-p", "3000", "-t", "20s", "-n", "a,b", "-v"})
	require.NoError(t, err)

	assert.Equal(t, "global.yaml", cfg)
	assert.Equal(t, "global.yaml", String("config"))
	assert.True(t, watch)
	assert.True(t, Bool("watch"))
	assert.Equal(t, 3000, port)
	assert.Equal(t, 3000, Int("port"))
	assert.Equal(t, 20*time.Second, dur)
	assert.Equal(t, 20*time.Second, Duration("timeout"))
	assert.Equal(t, []string{"a", "b"}, nodes)
	assert.Equal(t, []string{"a", "b"}, StringSlice("nodes"))
	assert.True(t, Bool("verbose"))

	assert.True(t, Changed("config"))
	assert.True(t, Changed("verbose"))
	assert.True(t, Has("config"))
	assert.False(t, Has("non-existent"))
	assert.NotNil(t, Lookup("config"))
	assert.NotNil(t, RootCommand())
	assert.NotNil(t, GlobalFlagSet())

	SetRootCommand(&cobra.Command{Use: "custom_root"})
	assert.Equal(t, "custom_root", RootCommand().Use)

	AddCommand(&cobra.Command{Use: "version"})
	assert.NotNil(t, RootCommand())
}

func TestFlag_UndefinedErrors(t *testing.T) {
	fs := NewFlagSet("errors")
	_ = fs.ParseArgs([]string{})

	_, err := fs.GetString("unknown")
	assert.Error(t, err)

	_, err = fs.GetBool("unknown")
	assert.Error(t, err)

	_, err = fs.GetInt("unknown")
	assert.Error(t, err)

	_, err = fs.GetInt64("unknown")
	assert.Error(t, err)

	_, err = fs.GetUint("unknown")
	assert.Error(t, err)

	_, err = fs.GetUint64("unknown")
	assert.Error(t, err)

	_, err = fs.GetFloat64("unknown")
	assert.Error(t, err)

	_, err = fs.GetDuration("unknown")
	assert.Error(t, err)

	// Safe getters return zero-values
	assert.Equal(t, "", fs.String("unknown"))
	assert.Equal(t, false, fs.Bool("unknown"))
	assert.Equal(t, 0, fs.Int("unknown"))
	assert.Equal(t, int64(0), fs.Int64("unknown"))
	assert.Equal(t, uint(0), fs.Uint("unknown"))
	assert.Equal(t, uint64(0), fs.Uint64("unknown"))
	assert.Equal(t, 0.0, fs.Float64("unknown"))
	assert.Equal(t, time.Duration(0), fs.Duration("unknown"))
	assert.Empty(t, fs.StringSlice("unknown"))
	assert.Empty(t, fs.IntSlice("unknown"))
}
