//go:build mage

package main

import (
	"context"

	"github.com/magefile/mage/mg"

	//mage:import
	"github.com/dosquad/mage"
)

// Local update, protoc, format, tidy, lint & test.
func Local(ctx context.Context) {
	mg.SerialCtxDeps(ctx, mage.Golang.Lint)
	mg.SerialCtxDeps(ctx, mage.Golang.Test)
	mg.SerialCtxDeps(ctx, mage.Goreleaser.Healthcheck)
	mg.SerialCtxDeps(ctx, mage.Goreleaser.Lint)
	mg.SerialCtxDeps(ctx, Regression)
	// mg.CtxDeps(ctx, mage.Test)
}

var Default = Local
