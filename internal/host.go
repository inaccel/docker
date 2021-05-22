package internal

import (
	"net/url"
	"os"
	"syscall"
)

var Host *url.URL

func Rootless() (bool, error) {
	info, err := os.Stat(Host.Path)
	if err != nil {
		return false, err
	}
	return info.Sys().(*syscall.Stat_t).Uid != 0, nil
}
