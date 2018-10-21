package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/k0kubun/pp"
	"github.com/mitchellh/cli"
)

type addCommand struct{}

func (c *addCommand) Help() string {
	return "Usage: dept add <url>"
}

func (c *addCommand) Synopsis() string {
	return "Add new CLI tool as a dependency"
}

func (c *addCommand) Run(args []string) int {
	pp.Println(args)
	if len(args) != 1 {
		fmt.Println(c.Help())
		return 1
	}

	ctx := context.Background()

	path := args[0]

	dir, err := filepath.Abs("./vendor")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("mkdir: %s\n", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalln(err)
	}

	newGOPATH := dir
	fmt.Printf("set GOPATH to vendor: %s\n", newGOPATH)
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", newGOPATH)
	defer os.Setenv("GOPATH", gopath)

	// 依存をすべて取得し、対象リポジトリで `dep ensure` する
	fmt.Printf("go get repository...: %s\n", path)
	b, err := exec.CommandContext(ctx, "go", "get", "-d", path).CombinedOutput()
	fmt.Println(string(b))
	if err != nil {
		log.Fatalln(err)
	}

	dist := filepath.Join("vendor", "src", path)
	fmt.Printf("change directory: %s\n", dist)
	os.Chdir(dist)

	if _, err := os.Stat("Gopkg.toml"); os.IsNotExist(err) {
		log.Fatalln("Gopkg.toml not found, abort")
	}

	fmt.Printf("dep ensure: %s\n", dist)
	if err := exec.CommandContext(ctx, "dep", "ensure").Run(); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("go install", path)
	b, err = exec.CommandContext(ctx, "go", "install", path).CombinedOutput()
	fmt.Println(string(b))
	if err != nil {
		log.Fatalln(err)
	}

	b, err = exec.CommandContext(ctx, "git", "show").Output()
	if err != nil {
		log.Fatalln(err)
	}
	digest := fmt.Sprintf("%x", sha256.New().Sum(b))

	b, err = exec.CommandContext(ctx, "git", "rev-parse", "HEAD").Output()
	if err != nil {
		log.Fatalln(err)
	}
	rev := string(b)

	b, err = json.MarshalIndent(&dependency{
		Digest:   digest,
		Revision: rev,
	}, "", "  ")
	io.WriteString(os.Stdout, string(b))

	return 0
}

func Add() (cli.Command, error) {
	return &addCommand{}, nil
}

type dependency struct {
	Revision string `json:"revision"`
	Digest   string `json:"digest"`
}
