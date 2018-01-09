package xdgbasedir

import (
	"os"
	"os/user"
	"testing"
)

func resetEnv() {
	// Reset overridden functions that were changed during testing
	osGetEnv = os.Getenv
	userCurrent = user.Current
	osStat = os.Stat
	osIsNotExist = os.IsNotExist
}

func expectEquals(t *testing.T, expected string, given string, msg string) {
	if expected != given {
		t.Error("Expected", expected, "not equal to given", given, ":", msg)
	}
}

// clearEnvVars explicitly unsets all XDG environment variables, to ensure that
// user environment does not affect tests
func clearEnvVars() {
	// explicitly unset, in case set in environment
	os.Setenv("HOME", "")
	os.Setenv("XDG_DATA_HOME", "")
	os.Setenv("XDG_CONFIG_HOME", "")
	os.Setenv("XDG_CACHE_HOME", "")
	os.Setenv("XDG_CONFIG_DIRS", "")
	os.Setenv("XDG_DATA_DIRS", "")
	os.Setenv("XDG_RUNTIME_DIR", "")
}

// TestGetAll checks path resolution when no XDG_* env vars nor HOME are set
func TestGetAll(t *testing.T) {
	clearEnvVars()

	// mock current user home directory location
	userCurrent = func() (*user.User, error) {
		return &user.User{HomeDir: "/home/Person"}, nil
	}
	defer resetEnv()

	dataHome, err := DataHomeDirectory()
	if err != nil {
		t.Error("Unexpected error ", err)
	}
	expectEquals(t, "/home/Person/.local/share", dataHome, "Unexpected data")

	configHome, err := ConfigHomeDirectory()
	if err != nil {
		t.Error("Unexpected error ", err)
	}
	expectEquals(t, "/home/Person/.config", configHome, "Unexpected config")

	cacheHome, err := CacheDirectory()
	if err != nil {
		t.Error("Unexpected error ", err)
	}
	expectEquals(t, "/home/Person/.cache", cacheHome, "Unexpected cache")
}

// TestDirectories
func TestDirectories(t *testing.T) {
	clearEnvVars()

	// mock current user home directory location
	userCurrent = func() (*user.User, error) {
		return &user.User{HomeDir: "/home/Person"}, nil
	}
	defer resetEnv()

	location, err := GetDataFileLocation("name")
	if err != nil {
		t.Error("Unexpected error ", err)
	}
	expectEquals(t, "/home/Person/.local/share/name", location, "Data file location")

	osGetEnv = func(string) string { return "/var/location/default" }

	location, err = GetDataFileLocation("name")
	if err != nil {
		t.Error("Unexpected error ", err)
	}
	expectEquals(t, "/var/location/default/name", location, "Data file location")
}
