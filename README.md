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

## Requirements
- Go v1.13 or later

## Basic usage
At first, let's create `gotool.mod` in a project root by the following command.
All tools which are managed by `dept` are written to `gotool.mod`.

``` sh
$ dept init
```

Then, let's install Go tools you want to use in your project.
``` sh
$ dept get github.com/mitchellh/gox github.com/tcnksm/ghr@v0.12.0
$ dept get -o lint github.com/golangci/golangci-lint/cmd/golangci-lint # rename golangci-lint as 'lint'
```

Finally, use `exec` to execute installed commands.
``` sh
$ dept exec ghr -v
ghr version v0.12.0
```

If you want to installed commands without `dept`, please run `build`.
``` sh
$ dept build
$ ls _tools
ghr     gox     lint
```

## Available commands
### init
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

### exec
`dept exec` executes the passed tool with arguments.

``` sh
$ dept exec ghr -v
```

### build
`dept build` builds all tools.

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
$ dept list
github.com/golangci/golangci-lint/cmd/golangci-lint lint v1.12.3
github.com/mitchellh/gox gox v0.4.0
github.com/tcnksm/ghr ghr v0.12.0
```

You can format output with `-f` flag.
``` sh
$ dept list -f '{{ .Name }}'
lint
gox
ghr
```

### clean
`dept clean` cleans up all cached tools.

``` sh
$ dept clean
```
