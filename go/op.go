package main

import (
	"io/ioutil"
	"os"
)

func create(path, content string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		panic(err)
	}
}

func scan(path string) []os.FileInfo {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	return dir
}

func read(path string) []byte {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return contents
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
