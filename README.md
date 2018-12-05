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
At first, let's create `gotool.mod` in a project root by the following command.
``` sh
$ dept init
```

To add a new tool to the project:
``` sh
$ dept get github.com/mitchellh/gox
```
