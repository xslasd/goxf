package conf

import (
	"time"
)

var defaultConfiguration = NewConf()

var configPassword string
var configCryptFilePath string

func SetPassword(pwd string) {
	configPassword = pwd
}
func SetCryptFilePath(path string) {
	configCryptFilePath = path
}

func LoadFromDataSource(ds ConfigSource, unmarshal Unmarshal) error {
	return defaultConfiguration.LoadFromConfigSource(ds, unmarshal)
}

// SetWatcher set a func in a watcher with name of key
func SetWatcher(key string, handle func(*Conf)) {
	defaultConfiguration.SetWatcher(key, handle)
}

// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) any {
	return defaultConfiguration.get(key)
}

// GetString returns the value associated with the key as a string with default defaultConfiguration.
func GetString(key string) string {
	return defaultConfiguration.GetString(key)
}

// GetBool returns the value associated with the key as a boolean with default defaultConfiguration.
func GetBool(key string) bool {
	return defaultConfiguration.GetBool(key)
}

// GetInt returns the value associated with the key as an integer with default defaultConfiguration.
func GetInt(key string) int {
	return defaultConfiguration.GetInt(key)
}

// GetInt64 returns the value associated with the key as an integer with default defaultConfiguration.
func GetInt64(key string) int64 {
	return defaultConfiguration.GetInt64(key)
}

// GetFloat64 returns the value associated with the key as a float64 with default defaultConfiguration.
func GetFloat64(key string) float64 {
	return defaultConfiguration.GetFloat64(key)
}

// GetTime returns the value associated with the key as time with default defaultConfiguration.
func GetTime(key string) time.Time {
	return defaultConfiguration.GetTime(key)
}

// GetDuration returns the value associated with the key as a duration with default defaultConfiguration.
func GetDuration(key string) time.Duration {
	return defaultConfiguration.GetDuration(key)
}

// GetStringSlice returns the value associated with the key as a slice of strings with default defaultConfiguration.
func GetStringSlice(key string) []string {
	return defaultConfiguration.GetStringSlice(key)
}

// GetSlice returns the value associated with the key as a slice of strings with default defaultConfiguration.
func GetSlice(key string) []any {
	return defaultConfiguration.GetSlice(key)
}

// GetStringMap returns the value associated with the key as a map of interfaces with default defaultConfiguration.
func GetStringMap(key string) map[string]any {
	return defaultConfiguration.GetStringMap(key)
}

// GetStringMapString returns the value associated with the key as a map of strings with default defaultConfiguration.
func GetStringMapString(key string) map[string]string {
	return defaultConfiguration.GetStringMapString(key)
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings with default defaultConfiguration.
func GetStringMapStringSlice(key string) map[string][]string {
	return defaultConfiguration.GetStringMapStringSlice(key)
}

// UnmarshalKey takes a single key and unmarshal it into a Struct with default defaultConfiguration.
func UnmarshalKey(key string, rawVal any) error {
	return defaultConfiguration.UnmarshalKey(key, rawVal)
}
