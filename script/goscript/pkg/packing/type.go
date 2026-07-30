package packing

import "fmt"

type PackageType uint8

const (
	PackageDEB PackageType = iota
	PackageRPM
	PackageIPK
	PackageOpenWrtAPK
	PackageAlpineAPK
	PackageArchLinux
)

func (p PackageType) String() string {
	switch p {
	case PackageDEB:
		return "deb"
	case PackageRPM:
		return "rpm"
	case PackageIPK:
		return "ipk"
	case PackageOpenWrtAPK:
		return "openwrt.apk"
	case PackageAlpineAPK:
		return "alpine.apk"
	case PackageArchLinux:
		return "archlinux"
	default:
		return fmt.Sprintf("PackageType(%d)", p)
	}
}

func (p PackageType) Nfpm() string {
	switch p {
	case PackageOpenWrtAPK, PackageAlpineAPK:
		return "apk"
	default:
		return p.String()
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
