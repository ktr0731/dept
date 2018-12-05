# dept

[![CircleCI](https://circleci.com/gh/ktr0731/dept.svg?style=svg)](https://circleci.com/gh/ktr0731/dept)
[![codecov](https://codecov.io/gh/ktr0731/dept/branch/master/graph/badge.svg?token=GLDI0EuIJs)](https://codecov.io/gh/ktr0731/dept)  

[Go modules](//github.com/golang/go/wiki/Modules) based dependency management for Go tools.

## Description
`dept` is a dependency management tool based on [Go modules](//github.com/golang/go/wiki/Modules).  
Instead of `go.mod`, `dept` helps you to manage Go tools. 
Go tools like [Golint](https://github.com/golang/lint), [errcheck](https://github.com/kisielk/errcheck) are often used in various environment.
`dept` provides you deterministic builds by manage tool dependencies.

`dept` is based on Go modules. All dependency resolution are provided by `go mod` commands.

## Usage

### init
At first, let's create `gotool.mod` in a project root by the following command.
``` sh
$ dept init
```

### get
`dept get` installs binaries to the specified directory.

``` sh
$ dept get github.com/mitchellh/gox
```

You can select the specified version like Go modules:
``` sh
$ dept get github.com/mitchellh/gox@v0.3.0
$ dept get github.com/mitchellh/gox@v0.1.0
```

To install a binary with another name:
``` sh
$ dept get -o lint github.com/golangci-lint/cmd/golangci-lint
```

Update tools to the latest version:
``` sh
$ dept get -u github.com/mitchellh/gox
$ dept get -u # update all tools
```

### remove
`dept remove` uninstalls passed tools.

``` sh
$ dept remove github.com/mitchellh/gox
```

### build
`dept build` builds all managed tools.

``` sh
$ dept build
```

If `$GOBIN` enabled, it will be used preferentially.
``` sh
$ GOBIN=$PWD/bin dept build
```

Also, `-d` flag is provided.
``` sh
$ dept build -d bin
```

### list
`dept list` list ups all tools managed by `dept`.

``` sh
$ dept get github.com/golangci/golangci-lint/cmd/golangci-lint github.com/matryer/moq
$ dept list
github.com/golangci/golangci-lint/cmd/golangci-lint v1.12.3
github.com/matryer/moq v0.0.0-20181107154629-5df7c6ae5624
```
