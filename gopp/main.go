package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
)

func processFile(file string) {
	buf, err := ioutil.ReadFile(file)

	s := fmt.Sprintf("%s", buf)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered ", r)
		}
	}()

	s = ProcessString(s)

	s = "//** This file is code generated by gopp. Do not edit.\n\n\n" + s

	i := strings.LastIndex(file, ".")

	if i < 0 {
		file = file + ".go"
	} else {
		file = file[:i] + ".go"
	}
	ioutil.WriteFile(file, []byte(s), os.ModePerm)

	execCommand("go fmt " + file)
}

/**
Process a string that is gopp formatted code, and return the go code
*/
func ProcessString(input string) string {
	l := lex(string(input))

	tree := parse(l)

	s := tree.String()

	return s
}


// execCommand wraps exec.Command
func execCommand(command string) {
	parts := strings.Split(command, " ")
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	var all bool
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Usage: gopp [file ...] [-all]")
		fmt.Println("-all: process all .gpp files in the current directory")
	}

	flag.BoolVar(&all, "all", false, "a string var")

	flag.Parse()
	//var err error

	if all {
		// Process all files
		files, _ := filepath.Glob("*.gpp")
		if len(files) == 0 {
			fmt.Println("No .gpp files found in current directory.")
			return
		}
		for _, file := range files {
			processFile(file)
		}
	} else {
		for _, file := range args {
			processFile(file)
		}
	}
}