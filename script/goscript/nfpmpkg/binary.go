package nfpmpkg

import (
	"context"

	"github.com/duakc/lightddns/script/goscript/pkg/common"
	"github.com/duakc/lightddns/script/goscript/pkg/gobuild"
	"github.com/duakc/lightddns/script/goscript/pkg/packing"
	"github.com/duakc/lightddns/script/goscript/pkg/target"

	"github.com/goreleaser/nfpm/v2/files"
)

func BuildBinary(ctx context.Context, tgt target.Target, forPackager packing.PackageType) (string, error) {
	outputDir := common.BuildDraftDir("nfpm", "binary", forPackager.String())
	common.Infof("start build binary for %s , goos=%s, goarch=%s",
		forPackager.String(), tgt.GOOS, tgt.GOARCH)
	p := gobuild.DefaultParams()
	p.OutputDir = outputDir
	p.Qualified = true
	return gobuild.Binary(ctx, tgt, p)
}

func ContentForBinary(ctx context.Context, tgt target.Target, forPackager packing.PackageType) (files.Contents, error) {
	buildBinary, err := BuildBinary(ctx, tgt, forPackager)
	if err != nil {
		return nil, err
	}
	return ContentForBinaryPath(buildBinary), nil
}

func ContentForBinaryPath(path string) files.Contents {
	return files.Contents{
		{
			Source: path, Destination: "/usr/bin/lightddns",
			FileInfo: &files.ContentFileInfo{Mode: 0o755, Owner: "root", Group: "root"},
		},
	}
}
