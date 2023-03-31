// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// Package main defines a program that runs another program
// provided on standard input, prints its standard output, then
// terminates. It logs to stderr.
//
// The input is expected to be json content encoding an exec.Cmd
// structure extended with a boolean AppendToEnv field.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("runner: ")
	log.Print("starting")
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("read %q", in)
	var cmd struct {
		exec.Cmd
		AppendToEnv bool
	}
	if err := json.Unmarshal(in, &cmd); err != nil {
		log.Fatal(err)
	}
	if cmd.AppendToEnv {
		cmd.Env = append(os.Environ(), cmd.Env...)
	}
	log.Printf("cmd: %+v", cmd)
	out, err := cmd.Output()
	if err != nil {
		s := err.Error()
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			s += ": " + string(bytes.TrimSpace(eerr.Stderr))
		}
		log.Fatalf("%v failed with %s", cmd.Args, s)
	}
	if _, err := os.Stdout.Write(out); err != nil {
		log.Fatal(err)
	}
	log.Print("succeeded")
}
