// Copyright 2021 The X Authors.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/the-xlang/x/parser"
	"github.com/the-xlang/x/pkg/io"
	"github.com/the-xlang/x/pkg/x"
)

func help(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	helpContent := [][]string{
		{"help", "Show help."},
		{"version", "Show version."},
		{"init", "Initialize new project here."},
	}
	maxlen := len(helpContent[0][0])
	for _, part := range helpContent {
		length := len(part[0])
		if length > maxlen {
			maxlen = length
		}
	}
	var sb strings.Builder
	const space = 5 // Space of between command name and description.
	for _, part := range helpContent {
		sb.WriteString(part[0])
		mlchc := (maxlen - len(part[0])) + space
		for mlchc > 0 {
			sb.WriteByte(' ')
			mlchc--
		}
		sb.WriteString(part[1])
		sb.WriteByte('\n')
	}
	println(sb.String()[:sb.Len()-1])
}

func version(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	println("The X Programming Language\n" + x.Version)
}

func initProject(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	err := os.WriteFile(x.SettingsFile, []byte(`out_name main
cxx_out_dir ./
cxx_out_name x.cpp`), 0606)
	if err != nil {
		println(err.Error())
		return
	}
	println("Initialized project.")
}

func processCommand(namespace, cmd string) bool {
	switch namespace {
	case "help":
		help(cmd)
	case "version":
		version(cmd)
	case "init":
		initProject(cmd)
	default:
		return false
	}
	return true
}

func init() {
	x.ExecutablePath = filepath.Dir(os.Args[0])
	// Not started with arguments.
	if len(os.Args) < 2 {
		os.Exit(0)
	}
	var sb strings.Builder
	for _, arg := range os.Args[1:] {
		sb.WriteString(" " + arg)
	}
	os.Args[0] = sb.String()[1:]
	arg := os.Args[0]
	index := strings.Index(arg, " ")
	if index == -1 {
		index = len(arg)
	}
	if processCommand(arg[:index], arg[index:]) {
		os.Exit(0)
	}
}

func loadXSet() {
	// File check.
	info, err := os.Stat(x.SettingsFile)
	if err != nil || info.IsDir() {
		println(`X settings file ("` + x.SettingsFile + `") is not found!`)
		os.Exit(0)
	}
	x.XSettings = x.NewXSet()
	bytes, err := os.ReadFile(x.SettingsFile)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	errors := x.XSettings.Parse(bytes)
	if errors != nil {
		println("X settings has errors;")
		for _, err := range errors {
			println(err.Error())
		}
		os.Exit(0)
	}
}

func printErrors(errors []string) {
	defer os.Exit(0)
	for _, message := range errors {
		fmt.Println(message)
	}
}

var routines *sync.WaitGroup

func main() {
	f, err := io.GetX(os.Args[0])
	if err != nil {
		println(err.Error())
		return
	}
	loadXSet()
	routines = new(sync.WaitGroup)
	info := new(parser.ParseFileInfo)
	info.File = f
	info.Routines = routines
	routines.Add(1)
	go parser.ParseFile(info)
	routines.Wait()
	if info.Errors != nil {
		printErrors(info.Errors)
	}
	os.WriteFile(
		filepath.Join(
			x.XSettings.Fields["cxx_out_dir"],
			x.XSettings.Fields["cxx_out_name"]),
		[]byte(info.X_CXX), 0606)
}
