package xdg

import (
	"os"
	"path/filepath"
	"strconv"
)

var (
	CacheHome  string
	ConfigDirs []string
	ConfigHome string
	DataDirs   []string
	DataHome   string
	RuntimeDir string
	StateHome  string
)

func init() {
	if value := os.Getenv("XDG_CACHE_HOME"); value != "" {
		CacheHome = value
	} else {
		CacheHome = filepath.Join(os.Getenv("HOME"), ".cache")
	}

	if value := os.Getenv("XDG_CONFIG_DIRS"); value != "" {
		ConfigDirs = filepath.SplitList(value)
	} else {
		ConfigDirs = []string{
			filepath.Join("/", "etc", "xdg"),
		}
	}

	if value := os.Getenv("XDG_CONFIG_HOME"); value != "" {
		ConfigHome = value
	} else {
		ConfigHome = filepath.Join(os.Getenv("HOME"), ".config")
	}

	if value := os.Getenv("XDG_DATA_DIRS"); value != "" {
		DataDirs = filepath.SplitList(value)
	} else {
		DataDirs = []string{
			filepath.Join("/", "usr", "local", "share"),
			filepath.Join("/", "usr", "share"),
		}
	}

	if value := os.Getenv("XDG_DATA_HOME"); value != "" {
		DataHome = value
	} else {
		DataHome = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}

	if value := os.Getenv("XDG_RUNTIME_DIR"); value != "" {
		RuntimeDir = value
	} else {
		RuntimeDir = filepath.Join("/", "run", "user", strconv.Itoa(os.Getuid()))
	}

	if value := os.Getenv("XDG_STATE_HOME"); value != "" {
		StateHome = value
	} else {
		StateHome = filepath.Join(os.Getenv("HOME"), ".local", "state")
	}
}
