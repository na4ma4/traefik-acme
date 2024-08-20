package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/dosquad/mage"
	"github.com/dosquad/mage/helper"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/na4ma4/go-permbits"
	"github.com/princjef/mageutil/shellcmd"
)

var errTestFailed = errors.New("test failed")

func runDebugCommand(_ context.Context, title, acmeFile, path string, certMatch, keyMatch []byte, args []string) error {
	args = append([]string{
		// --acme="$(<)"
		"--acme=" + acmeFile,
		// --cert "$(@D)/cert.pem"
		"--cert=" + filepath.Join(path, "cert.pem"),
		// --key "$(@D)/key.pem"
		"--key=" + filepath.Join(path, "key.pem"),
	}, args...)

	sh.Rm(path)
	helper.MustMakeDir(path, permbits.MustString("ug=rwx,o=rx"))

	cmdName := helper.GetEnv("RUN_CMD", helper.Must[string](helper.FirstCommandName()))
	ct := helper.NewCommandTemplate(true, "./cmd/"+cmdName)
	{
		err := shellcmd.Command(fmt.Sprintf("%s %s", ct.OutputArtifact, strings.Join(args, " "))).Run()
		if err != nil {
			return fmt.Errorf("%w: command execute failed: %w", errTestFailed, err)
		}
	}

	// mg.SerialCtxDeps(ctx, mg.F(
	// 	mage.RunE,
	// 	strings.Join(append([]string{}, args...), " "),
	// ))

	if len(certMatch) != 0 {
		if !helper.FileExists(path, "cert.pem") {
			return fmt.Errorf("%w: certificate file does not exist", errTestFailed)
		}
		body, err := os.ReadFile(filepath.Join(path, "cert.pem"))
		if err != nil {
			return fmt.Errorf("%w: unable to read certificate: %w", errTestFailed, err)
		}
		if !bytes.Contains(body, certMatch) {
			return fmt.Errorf("%w: certificate file does not contain '%s'", errTestFailed, certMatch)
		}
	}

	if len(keyMatch) != 0 {
		if !helper.FileExists(path, "key.pem") {
			return fmt.Errorf("%w: key file does not exist", errTestFailed)
		}
		body, err := os.ReadFile(filepath.Join(path, "key.pem"))
		if err != nil {
			return fmt.Errorf("%w: unable to read key: %w", errTestFailed, err)
		}
		if !bytes.Contains(body, keyMatch) {
			return fmt.Errorf("%w: key file does not contain '%s'", errTestFailed, keyMatch)
		}
	}

	helper.PrintInfo("Test Passed: %s", title)

	return nil
}

func regressionTestIssue5Part1(ctx context.Context) error {
	return runDebugCommand(
		ctx,
		"regression test issue5/1",
		"testdata/issue-5/acme.json",
		"artifacts/test/issue-5",
		[]byte("certificate-for-example.com"),
		[]byte("key-for-example.com"),
		[]string{
			`--certificate-resolver="acme-different"`,
			`*.example.com`,
		},
	)
}

func regressionTestIssue5Part2(ctx context.Context) error {
	return runDebugCommand(
		ctx,
		"regression test issue5/2",
		"testdata/issue-5/acme.json",
		"artifacts/test/issue-5",
		[]byte("certificate-for-example.com"),
		[]byte("key-for-example.com"),
		[]string{
			`--certificate-resolver="acme-different"`,
			`example.com`,
		},
	)
}

func regressionTestIssue14V1(ctx context.Context) error {
	return runDebugCommand(
		ctx,
		"regression test issue14/v1",
		"testdata/issue-14/v1/acme.json",
		"artifacts/test/issue-14/v1",
		[]byte("Certificate Public"),
		[]byte("Certificate Key"),
		[]string{
			`test.example.com`,
		},
	)
}

func regressionTestIssue14V2(ctx context.Context) error {
	return runDebugCommand(
		ctx,
		"regression test issue14/v2",
		"testdata/issue-14/v2/new-acme.json",
		"artifacts/test/issue-14/v2",
		[]byte("Certificate Public"),
		[]byte("Certificate Key"),
		[]string{
			`--certificate-resolver "myresolver"`,
			`test.example.com`,
		},
	)
}

func regressionTestIssue52(ctx context.Context) error {
	var u *user.User
	{
		var err error
		u, err = user.Current()
		if err != nil {
			return fmt.Errorf("%w: unable to get current user", err)
		}
	}

	var cfg *helper.DockerConfig
	{
		var err error
		cfg, err = helper.DockerLoadConfig()
		if err != nil {
			return fmt.Errorf("%w: unable to load docker configuration", err)
		}
	}

	sh.Rm(helper.MustGetArtifactPath("test", "issue-52"))
	helper.MustMakeDir(
		helper.MustGetArtifactPath("test", "issue-52"),
		permbits.MustString("ug=rwx,o=rx"),
	)

	mg.SerialCtxDeps(ctx, mage.Docker.Build)

	out, err := helper.CommandString(
		`docker run -t --rm ` +
			`--user ` + u.Uid + `:` + u.Gid + ` ` +
			`-v ` + helper.MustGetWD("testdata", "issue-52") + `:/input ` +
			`-v ` + helper.MustGetArtifactPath("test", "issue-52") + `:/output ` +
			`--workdir /output ` +
			cfg.GetImageRef() + ` ` +
			`--debug --acme /input/acme.json test.example.com`,
	)
	if err != nil {
		return fmt.Errorf("%w: unable to execute command", err)
	}

	helper.PrintDebug("Command Output\n%s", out)

	if !helper.FileExists(helper.MustGetArtifactPath("test", "issue-52", "cert.pem")) {
		return fmt.Errorf("%w: certificate file does not exist", errTestFailed)
	}
	body, err := os.ReadFile(helper.MustGetArtifactPath("test", "issue-52", "cert.pem"))
	if err != nil {
		return fmt.Errorf("%w: unable to read certificate: %w", errTestFailed, err)
	}
	if !bytes.Contains(body, []byte("Certificate")) {
		return fmt.Errorf("%w: certificate file does not contain 'Certificate'", errTestFailed)
	}

	helper.PrintInfo("Test Passed: regression test issue52")

	return nil
}

func Regression(ctx context.Context) {
	mg.SerialCtxDeps(ctx,
		mage.Build.Debug,
		regressionTestIssue5Part1,
		regressionTestIssue5Part2,
		regressionTestIssue14V1,
		regressionTestIssue14V2,
		regressionTestIssue52,
	)
}
