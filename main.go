package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: dept <url>")
	}

	ctx := context.Background()

	sp := strings.Split(os.Args[1], " ")
	path := filepath.Join(append([]string{"vendor"}, sp...)...)
	fmt.Printf("mkdir: %s\n", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatalln(err)
	}

	newGOPATH, err := filepath.Abs("./vendor")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("set GOPATH to vendor: %s\n", newGOPATH)
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", newGOPATH)
	defer os.Setenv("GOPATH", gopath)

	// 依存をすべて取得し、対象リポジトリで `dep ensure` する
	fmt.Printf("go get repository...: %s\n", os.Args[1])
	b, err := exec.CommandContext(ctx, "go", "get", "-d", os.Args[1]).CombinedOutput()
	fmt.Println(string(b))
	if err != nil {
		log.Fatalln(err)
	}

	dist := fmt.Sprintf("vendor/%s", os.Args[1])
	fmt.Printf("dep ensure: %s\n", dist)
	os.Chdir(dist)

	exec.CommandContext(ctx, "dep", "ensure")

	// fmt.Println("go install", os.Args[1])
	// b, err = exec.CommandContext(ctx, "go", "install", os.Args[1]).CombinedOutput()
	// fmt.Println(string(b))
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}
