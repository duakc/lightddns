// Package nfpmbuild holds the small mechanical steps the nfpm-based package
// builders (deb, rpm, archlinux, openwrt) share: staging the rendered config
// and gzipped man page, the systemd content bundle, and writing a package.
// It does not know about any single format's metadata or scripts - each builder
// owns those - so the formats stay independent.
package nfpmbuild

import (
	"fmt"
	"os"
	"path/filepath"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"
	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"
)

const subSchemaURL = "SCHEMA_URL"

var releaseDir = mt.Must(filehelper.New(common.ReleaseDir()))

// SchemaURL pins the config schema to the build version so an installed config
// validates against the matching schema.
func SchemaURL(version string) string {
	return fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/release/schema.json",
		constpkg.Repo, version)
}

// StageConfig renders the schema URL into the example config under dir and
// returns the rendered file's path.
func StageConfig(dir, schemaURL string) (string, error) {
	if err := render(dir, []packing.File{{
		FS: releaseDir, From: "example/lightddns.yaml", To: "lightddns.yaml",
		Mode: 0o640, SubSetVec: packing.SubSetVec{subSchemaURL},
	}}, packing.SubSet{subSchemaURL: schemaURL}); err != nil {
		return "", err
	}
	return filepath.Join(dir, "lightddns.yaml"), nil
}

// StageMan gzips the man page under dir and returns its path.
func StageMan(dir string) (string, error) {
	if err := render(dir, []packing.File{{
		FS: releaseDir, From: "man/lightddns.1", To: "lightddns.1.gz", Mode: 0o644, Gzip: true,
	}}, packing.SubSet{}); err != nil {
		return "", err
	}
	return filepath.Join(dir, "lightddns.1.gz"), nil
}

func render(dir string, fileList []packing.File, sub packing.SubSet) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fh, err := filehelper.New(dir)
	if err != nil {
		return err
	}
	defer fh.Close()
	return packing.ProcessAll(fh, fileList, sub)
}

// Binary is the content entry for the compiled binary at /usr/bin/lightddns.
func Binary(binPath string) *files.Content {
	return &files.Content{
		Source: binPath, Destination: "/usr/bin/lightddns",
		FileInfo: &files.ContentFileInfo{Mode: 0o755},
	}
}

// SystemdContents is the file set shared by the systemd-based packages (deb,
// rpm, archlinux): the three config files, both units, and the man page. The
// per-target binary is added by the caller.
func SystemdContents(configPath, manPath string) files.Contents {
	config := func(dst string) *files.Content {
		return &files.Content{Source: configPath, Destination: dst, Type: "config|noreplace", FileInfo: &files.ContentFileInfo{Mode: 0o640}}
	}
	unit := func(name string) *files.Content {
		return &files.Content{Source: common.ReleaseDir("systemd", name), Destination: "/usr/lib/systemd/system/" + name, FileInfo: &files.ContentFileInfo{Mode: 0o644}}
	}
	return files.Contents{
		config("/etc/lightddns.yaml"),
		config("/etc/lightddns.d/example.yaml"),
		{Source: common.ReleaseDir("example", "environment"), Destination: "/etc/default/lightddns", Type: "config|noreplace", FileInfo: &files.ContentFileInfo{Mode: 0o640}},
		unit("lightddns.service"),
		unit("lightddns@.service"),
		{Source: manPath, Destination: "/usr/share/man/man1/lightddns.1.gz", FileInfo: &files.ContentFileInfo{Mode: 0o644}},
	}
}

// WriteTo builds info as the given nfpm format and writes it to outPath.
func WriteTo(info *nfpm.Info, format, outPath string) error {
	pkgr, err := nfpm.Get(format)
	if err != nil {
		return err
	}
	info = nfpm.WithDefaults(info)
	if err := nfpm.Validate(info); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return pkgr.Package(info, f)
}
