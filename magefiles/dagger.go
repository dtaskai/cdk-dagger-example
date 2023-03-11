//go:build mage

package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

const (
	eslintVersion = "8"
	nodeVersion   = "18"
	awsCliVersion = "2.11.1"
	alpineVersion = "3.17.2"
)

func Deploy(ctx context.Context) {
	d := daggerClient(ctx)
	defer d.Close()

	deploy(ctx, d)
}

func deploy(ctx context.Context, d *dagger.Client) {
	accessKey := hostEnv(ctx, d.Host(), "AWS_ACCESS_KEY_ID").Secret()
	secretKey := hostEnv(ctx, d.Host(), "AWS_SECRET_ACCESS_KEY").Secret()
	defaultRegion := hostEnv(ctx, d.Host(), "AWS_DEFAULT_REGION").Secret()
	account := hostEnv(ctx, d.Host(), "CDK_DEFAULT_ACCOUNT").Secret()

	lint := lint(ctx, d)

	_, err := d.Container().
		From(fmt.Sprintf("node:%v", nodeVersion)).
		WithSecretVariable("AWS_ACCESS_KEY_ID", accessKey).
		WithSecretVariable("AWS_SECRET_ACCESS_KEY", secretKey).
		WithSecretVariable("AWS_DEFAULT_REGION", defaultRegion).
		WithSecretVariable("CDK_DEFAULT_REGION", defaultRegion).
		WithSecretVariable("CDK_DEFAULT_ACCOUNT", account).
		WithMountedDirectory("/build", lint).
		WithWorkdir("/build").
		WithExec([]string{"npm", "install", "-g", "aws-cdk"}).
		WithExec([]string{"npm", "install"}).
		WithExec([]string{"npm", "run", "build"}).
		WithExec([]string{"cdk", "deploy", "--require-approval", "never"}).
		ExitCode(ctx)

	if err != nil {
		fmt.Println(err)
	}
}

func Lint(ctx context.Context) {
	d := daggerClient(ctx)
	defer d.Close()

	lint(ctx, d)
}

func lint(ctx context.Context, d *dagger.Client) *dagger.Directory {
	install := install(ctx, d)

	lint := d.Container().
		From(fmt.Sprintf("cytopia/eslint:%v", eslintVersion)).
		WithMountedDirectory("/data", install).
		WithExec([]string{"."})

	_, err := lint.ExitCode(ctx)

	if err != nil {
		panic(unavailableErr(err))
	}

	return lint.Directory("/data")
}

func Install(ctx context.Context) {
	d := daggerClient(ctx)
	defer d.Close()

	install(ctx, d)
}

func install(ctx context.Context, d *dagger.Client) *dagger.Directory {
	install := d.Container().
		From(fmt.Sprintf("node:%v", nodeVersion)).
		WithMountedDirectory("/src", sourceCode(d)).
		WithWorkdir("/src").
		WithExec([]string{"npm", "ci", "&&", "npm", "run", "build"})

	_, err := install.ExitCode(ctx)

	if err != nil {
		panic(unavailableErr(err))
	}

	return install.Directory("/src")
}

func sourceCode(d *dagger.Client) *dagger.Directory {
	return d.Host().Directory(".")
}

func daggerClient(ctx context.Context) *dagger.Client {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(unavailableErr(err))
	}
	return client
}

func unavailableErr(err error) Exit {
	return Exit{Code: 69, Error: err}
}

type Exit struct {
	Code  int
	Error error
}

func hostEnv(ctx context.Context, host *dagger.Host, varName string) *dagger.HostVariable {
	hostEnv := host.EnvVariable(varName)
	return hostEnv
}
