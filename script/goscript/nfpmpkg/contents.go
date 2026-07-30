package nfpmpkg

import (
	"context"
	"slices"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/goreleaser/nfpm/v2/files"
)

type FileContents struct {
	files.Contents
}

func (fc *FileContents) append(cc files.Contents) *FileContents {
	fc.Contents = append(fc.Contents, cc...)
	return fc
}

func (fc *FileContents) AddConfig(schemaURL string) *FileContents {
	cc, err := ContentForConfigFiles(schemaURL)
	if err != nil {
		common.Fatalf("%s", err)
	}
	return fc.append(cc)
}

func (fc *FileContents) AddSystemdService() *FileContents {
	return fc.append(ContentForSystemdServices())
}

func (fc *FileContents) AddSystemdTmpFiles() *FileContents {
	return fc.append(ContentForSystemdTmpFiles())
}

func (fc *FileContents) AddSystemdSysUsers() *FileContents {
	return fc.append(ContentForSystemdUsers())
}

func (fc *FileContents) AddMan() *FileContents {
	return fc.append(ContentForMan())
}

func (fc *FileContents) AddEnvFile() *FileContents {
	return fc.append(ContentForEnvFile())
}

func (fc *FileContents) Copy() *FileContents {
	ffc := *fc
	ffc.Contents = slices.Clone(fc.Contents)
	return &ffc
}

func (fc *FileContents) AddBinary(ctx context.Context, tgt target.Target, packageType packing.PackageType) *FileContents {
	cc, err := ContentForBinary(ctx, tgt, packageType)
	if err != nil {
		common.Fatalf("%s", err)
	}
	return fc.append(cc)
}
