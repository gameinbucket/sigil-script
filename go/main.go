package main

//
//           The Hymn Compiler
// Copyright 2019 Nathan Michael McMillan
//

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	debug       = true
	debugTokens = false
	debugTree   = true
)

const (
	spaceChar = '\t'
	spaceFmc  = string(spaceChar)
)

func fmc(depth int) string {
	space := ""
	for i := 0; i < depth; i++ {
		space += spaceFmc
	}
	return space
}

func help() {
	fmt.Println("Hymn command line interface.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("    hymn <command> [arguments]")
	fmt.Println("")
	fmt.Println("The commands are:")
	fmt.Println("")
	fmt.Println("    fmt      format a file")
	fmt.Println("    build    compile a program")
}

func execFormat(args []string) {
	size := len(args)
	if size <= 2 {
		fmt.Println("[path]")
	} else {
		path := args[2]
		hymnFmt(path)
	}
}

func execBuild(args []string) {
	size := len(args)
	if size <= 3 {
		fmt.Println("[library] [path]")
		return
	}
	libDir := args[2]
	path := args[3]
	isLib := false
	if size >= 5 {
		if args[4] == "--lib" {
			isLib = true
		}
	}
	execCompile("out", path, libDir, isLib)
}

func main() {
	args := os.Args
	size := len(args)
	if size <= 1 {
		help()
	} else if args[1] == "fmt" {
		execFormat(args)
	} else if args[1] == "build" {
		execBuild(args)
	} else {
		help()
	}
}

func execCompile(out, path, libDir string, isLib bool) string {
	program := programInit()
	program.out = out
	program.libDir = libDir
	program.directory = fileDir(path)

	hmlib := &hmlib{}
	hmlib.libs()
	program.hmlib = hmlib

	program.compile(out, path, libDir)

	name := fileName(path)
	fileOut := out + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	gcc(program.sources, fileOut, isLib)
	return app(fileOut)
}

func (me *program) compile(out, path, libDir string) {
	name := fileName(path)
	hymn := me.hymnFileInit(name)
	me.hmfiles[name] = hymn
	me.hmorder = append(me.hmorder, hymn)
	hymn.parse(out, path)
	source := hymn.generateC(out, name, libDir)
	me.sources[name] = source
}

func gcc(sources map[string]string, fileOut string, isLib bool) {
	fmt.Println("=== gcc ===")
	paramGcc := make([]string, 0)
	paramGcc = append(paramGcc, "-Wall")
	paramGcc = append(paramGcc, "-Wextra")
	paramGcc = append(paramGcc, "-pedantic")
	paramGcc = append(paramGcc, "-std=c11")
	for _, src := range sources {
		paramGcc = append(paramGcc, src)
	}
	paramGcc = append(paramGcc, "-o")
	if isLib {
		fileOut += ".o"
		paramGcc = append(paramGcc, fileOut)
		paramGcc = append(paramGcc, "-c")
	} else {
		paramGcc = append(paramGcc, fileOut)
	}
	fmt.Println("gcc", strings.Join(paramGcc, " "))
	cmd := exec.Command("gcc", paramGcc...)
	stdout, err := cmd.CombinedOutput()
	std := string(stdout)
	if std != "" {
		fmt.Println(std)
	}
	if err != nil {
		panic(err)
	}
}

func app(path string) string {
	if exists(path) {
		fmt.Println("=== run ===")
		stdout, _ := exec.Command(path).CombinedOutput()
		finalout := string(stdout)
		fmt.Println(finalout)
		return finalout
	}
	fmt.Println("===")
	return ""
}
