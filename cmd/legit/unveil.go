//go:build openbsd
// +build openbsd

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

func Unveil(path string, perms string) error {
	log.Printf("unveil: \"%s\", %s", path, perms)
	return unix.Unveil(path, perms)
}

func UnveilBlock() error {
	log.Printf("unveil: block")
	return unix.UnveilBlock()
}

func UnveilPaths(paths []string, perms string) error {
	for _, path := range paths {
		if err := Unveil(path, perms); err != nil {
			return err
		}
	}
	return UnveilBlock()
}
