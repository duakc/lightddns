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
	return gobuild.Plain(ctx, tgt, outputDir)
}

func ContentForBinary(ctx context.Context, tgt target.Target, forPackager packing.PackageType) (files.Contents, error) {
	buildBinary, err := BuildBinary(ctx, tgt, forPackager)
	if err != nil {
		return nil, err
	}
	return files.Contents{
		{Source: buildBinary, Destination: "/usr/bin/lightddns",
			FileInfo: &files.ContentFileInfo{Mode: 0755, Owner: "root", Group: "root"}},
	}, nil
}
