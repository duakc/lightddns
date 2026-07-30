package packing

import "fmt"

type PackageType uint8

const (
	PackageDEB PackageType = iota
	PackageRPM
	PackageAPK
	PackageIPK
	PackageArchLinux
)

func (p PackageType) String() string {
	switch p {
	case PackageDEB:
		return "deb"
	case PackageRPM:
		return "rpm"
	case PackageAPK:
		return "apk"
	case PackageIPK:
		return "ipk"
	case PackageArchLinux:
		return "archlinux"
	default:
		return fmt.Sprintf("PackageType(%d)", p)
	}
}

func (p PackageType) Ext() string {
	switch p {
	case PackageArchLinux:
		return "pkg.zst"
	default:
		return p.String()
	}
}
