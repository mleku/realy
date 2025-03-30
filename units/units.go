// Package units is a convenient set of names designating data sizes in bytes
// using common ISO names (base 10).
package units

const (
	Kilobyte = 1000
	Kb       = Kilobyte
	Megabyte = Kilobyte * Kilobyte
	Mb       = Megabyte
	Gigabyte = Megabyte * Kilobyte
	Gb       = Gigabyte
)
