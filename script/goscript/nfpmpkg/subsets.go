package nfpmpkg

import (
	"fmt"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"

	"github.com/goreleaser/nfpm/v2/files"
)

const (
	SubSetSchemaURL = "SCHEMA_URL"
)

var draftDirFileHelper = mt.Must(
	filehelper.New(common.BuildDraftDir("nfpm")),
)

var releaseDirFileHelper = mt.Must(
	filehelper.New(common.ReleaseDir()),
)

func ContentForConfigFiles(schemaURL string) (files.Contents, error) {
	const configFileExample = "example/lightddns.yaml"
	configFile := packing.File{
		FS: releaseDirFileHelper, From: configFileExample,
		SubSetVec: packing.BuildSubSetVec(SubSetSchemaURL),
	}
	renderFilePath, err := configFile.Render(draftDirFileHelper, packing.SubSet{
		SubSetSchemaURL: schemaURL,
	})
	if err != nil {
		return nil, err
	}
	return files.Contents{
		{Source: renderFilePath, Destination: "/etc/lightddns.yaml", Type: files.TypeConfigNoReplace},
		{Source: renderFilePath, Destination: "/etc/lightddns.d/example.yaml", Type: files.TypeConfigNoReplace},
	}, nil
}

func ContentForSystemdServices() files.Contents {
	return files.Contents{
		{
			Source:      releaseDirFileHelper.Path("systemd", "lightddns.service"),
			Destination: "/usr/lib/systemd/system/lightddns.service",
			FileInfo:    &files.ContentFileInfo{Mode: 0o644},
		},
		{
			Source:      releaseDirFileHelper.Path("systemd", "lightddns@.service"),
			Destination: "/usr/lib/systemd/system/lightddns@.service",
			FileInfo:    &files.ContentFileInfo{Mode: 0o644},
		},
	}
}

func ContentForOpenWrtInit() files.Contents {
	return files.Contents{
		{
			Source:      releaseDirFileHelper.Path("openwrt", "lightddns.init"),
			Destination: "/etc/init.d/lightddns",
			FileInfo:    &files.ContentFileInfo{Mode: 0o755},
		},
	}
}

func ContentForAlpineOpenRC() files.Contents {
	return files.Contents{
		{
			Source:      releaseDirFileHelper.Path("alpine", "lightddns.initd"),
			Destination: "/etc/init.d/lightddns",
			FileInfo:    &files.ContentFileInfo{Mode: 0o755},
		},
		{
			Source:      releaseDirFileHelper.Path("alpine", "lightddns.confd"),
			Destination: "/etc/conf.d/lightddns",
			Type:        files.TypeConfigNoReplace,
			FileInfo:    &files.ContentFileInfo{Mode: 0o644},
		},
	}
}

func ContentForEnvFile() files.Contents {
	return files.Contents{
		{
			Source:      releaseDirFileHelper.Path("example", "environment"),
			Destination: "/etc/default/lightddns", Type: files.TypeConfigNoReplace,
			FileInfo: &files.ContentFileInfo{Mode: 0o600},
		},
	}
}

func ContentForMan() files.Contents {
	configFile := packing.File{
		FS: releaseDirFileHelper, From: "man/lightddns.1",
		Gzip: true,
	}
	renderFilePath, err := configFile.Render(draftDirFileHelper, nil)
	if err != nil {
		common.Fatalf("rendering man page: %s", err)
	}
	return files.Contents{
		{
			Source:      renderFilePath,
			Destination: "/usr/share/man/man1/lightddns.1.gz",
			FileInfo:    &files.ContentFileInfo{Mode: 0o644},
		},
	}
}

func ContentForSystemdTmpFiles() files.Contents {
	return files.Contents{
		{
			Source:      releaseDirFileHelper.Path("systemd", "tmpfiles", "lightddns.conf"),
			Destination: "/usr/lib/tmpfiles.d/lightddns.conf",
			FileInfo:    &files.ContentFileInfo{Mode: 0o644},
		},
	}
}

func ContentForSystemdUsers() files.Contents {
	return files.Contents{
		{
			Source:      releaseDirFileHelper.Path("systemd", "sysusers", "lightddns.conf"),
			Destination: "/usr/lib/sysusers.d/lightddns.conf",
			FileInfo:    &files.ContentFileInfo{Mode: 0o644},
		},
	}
}

func SchemaURL(version string) string {
	return fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/release/schema.json",
		constpkg.Repo, version)
}
