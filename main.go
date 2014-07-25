package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}
	if err := filepath.Walk(dir, run); err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Analyzed %d directories with %d files.", dirs, files)
}

var (
	tag    = []byte("// +build !windows\n")
	prefix = []byte("// +build")
	dirs   = 1
	files  = 0
)

func run(pth string, info os.FileInfo, err error) error {
	if info.IsDir() {
		switch path.Base(pth) {
		case ".bzr", ".git", ".svn", ".hg":
			// don't navigate into these.
			return filepath.SkipDir
		}
		dirs++
		return nil
	}

	if !strings.HasSuffix(pth, "_test.go") {
		// only _test.go files
		return nil
	}

	files++

	b, err := ioutil.ReadFile(pth)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	if bytes.HasPrefix(b, prefix) {
		buildTag := bytes.SplitN(b, []byte("\n"), 2)[0]
		if !bytes.Equal(tag[:len(tag)-1], buildTag) {
			fmt.Printf("%s already has tag %s\n", pth, string(buildTag))
		}
		return nil
	}

	f, err := os.OpenFile(pth, os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %v", err)
	}
	defer f.Close()
	if _, err := f.Write(tag); err != nil {
		return err
	}
	_, err = f.Write(b)

	if err == nil {
		fmt.Printf("%s: Modified.\n", pth)
	}
	return err
}
