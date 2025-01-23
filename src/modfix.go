// Copyright (c) 2024 BlueRock Security, Inc.

// cSpell:ignore modfile
// cSpell:disable

package src

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/mod/modfile"
)

// cSpell:enable

func Execute(config *Config) (*Report, error) {
	r := &handle{Config: config}
	return r.init()
}

type handle struct {
	*Config
	goPath   map[string]*file
	workPath map[string]*file
}

type file struct {
	*handle
	*modfile.File
	filename   string
	inWorkPath bool
	old        []byte
	paths      map[string]bool
	updated    *bool
	err        error
	state      *ModFileState
}

func (x *file) report() *ReportEntry {
	return &ReportEntry{
		Filename: x.filename,
		Module:   x.Module.Mod.Path,
		Error:    x.err,
		State:    x.state,
		Updated:  x.updated,
	}
}

func (x *file) init() {

	x.paths = make(map[string]bool)

	for _, v := range x.Replace {
		if v != nil && v.Syntax != nil {
			x.DropReplace(v.Old.Path, v.Old.Version)
		}
	}

	err := x.addMod(x, true)
	if err != nil {
		x.err = err
		return
	}

	new, err := x.Format()
	if err != nil {
		x.err = err
		return
	}

	if bytes.Equal(x.old, new) {
		x.state = ModFileStateClean.Pointer()
	} else {
		x.state = ModFileStateDirty.Pointer()
		if !x.DryRun {
			err = os.WriteFile(x.filename, new, 0644)
			if err != nil {
				x.err = err
				return
			}
			xTrue := true
			x.updated = &xTrue
		} else {
			xFalse := false
			x.updated = &xFalse
		}

	}
}

func (x *file) addMod(mod *file, recursion bool) error {

	prepend := fmt.Sprintf("(%s) addMod(mod=%s)", x.Module.Mod.Path, mod.Module.Mod.Path)

	addReplace := func(oldPath string, newPath string) error {

		if x.paths[oldPath] {
			x.log("%s->already added", prepend)
			return nil
		}

		err := x.AddReplace(oldPath, "", newPath, "")
		if err != nil {
			return err
		}

		x.paths[oldPath] = true
		x.log("%s->added", prepend)

		return nil
	}

	relative, err := filepath.Rel(filepath.Dir(x.filename), filepath.Dir(mod.filename))
	if err != nil {
		panic(err)
	}

	if !strings.HasPrefix(relative, "../") {
		relative = "./" + relative
	}

	if x == mod {
		x.log("%s->not adding self", prepend)
	} else {
		addReplace(mod.Module.Mod.Path, relative)
	}

	if !recursion {
		return nil
	}

	var errs *multierror.Error

	if x.AddAll {

		x.log("%s->add all", prepend)

		for _, v := range x.goPath {

			err := x.addMod(v, false)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

		}
	} else {

		x.log("%s->add required only", prepend)

		for _, v := range mod.Require {

			child := x.goPath[v.Mod.Path]
			if child == nil {
				x.log("%s->%s not found", prepend, v.Mod.Path)
				continue
			}

			x.log("%s->%s found", prepend, v.Mod.Path)
			err := x.addMod(child, true)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

		}
	}

	return errs.ErrorOrNil()
}

func (x *handle) init() (*Report, error) {

	r := &Report{}

	prepend := "(handle) init()"

	x.goPath = make(map[string]*file)
	x.workPath = make(map[string]*file)

	if x.GoPath == "" {
		return nil, fmt.Errorf("%s->error gopath is required", prepend)
	}

	if x.WorkPath == "" {
		x.WorkPath = "."
	}

	err := func() error {

		var err error

		x.GoPath, err = filepath.Abs(x.GoPath)
		if err != nil {
			return err
		}

		x.WorkPath, err = filepath.Abs(x.WorkPath)
		if err != nil {
			return err
		}

		err = x.readDir(x.GoPath, false)
		if err != nil {
			return err
		}

		err = x.readDir(x.WorkPath, true)
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		return nil, err
	}

	workPathLen := len(x.workPath)

	if workPathLen == 0 {
		return nil, fmt.Errorf("no mod files found in work path")
	}

	for _, v := range x.workPath {
		v.init()
	}

	errCount := 0

	for _, v := range x.workPath {
		entry := v.report()
		if entry.Error != nil {
			errCount++
		}
		r.WorkPathFiles = append(r.WorkPathFiles, entry)
	}

	for _, v := range x.goPath {
		if x.Verbose {
			r.GoPathFiles = append(r.GoPathFiles, v.report())
		}
	}

	if errCount == 0 {
		r.State = ReportStateSuccess
		return r, nil
	}

	if errCount == workPathLen {
		r.State = ReportStateFail
		return r, nil
	}

	r.State = ReportStatePartial
	return r, nil
}

func (x *handle) readDir(directory string, inWorkPath bool) error {

	prepend := fmt.Sprintf("readDir(directory=%s, inWorkPath=%t)", directory, inWorkPath)

	directory, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("%s->error getting absolute path for directory %s", prepend, directory)
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("%s->error reading directory %s", prepend, directory)
	}

	addMod := func(filename string) error {

		b, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		m, err := modfile.Parse("go.mod", b, nil)
		if err != nil {
			return fmt.Errorf("%s->error parsing %s: %w", prepend, filename, err)
		}

		gm := &file{handle: x, File: m, filename: filename, old: b, inWorkPath: inWorkPath}

		x.goPath[m.Module.Mod.Path] = gm

		if inWorkPath {
			x.workPath[m.Module.Mod.Path] = gm
			return nil
		}

		return nil
	}

	var errs *multierror.Error

	for _, file := range files {

		filename := path.Join(directory, file.Name())

		if file.IsDir() {

			if inWorkPath {

				if !x.Recursive {
					x.log("%s->skipping directory %s (recursive=false)", prepend, file.Name())
					continue
				}

			} else {

				if strings.HasPrefix(file.Name(), ".") && !x.EnableHidden {
					x.log("%s->skipping hidden directory %s", prepend, file.Name())
					continue
				}

				if directory == x.GoPath {

					switch file.Name() {

					case "pkg", "bin":
						x.log("%s->skipping %s directory", prepend, file.Name())
						continue

					}

				}

			}

			err := x.readDir(filename, inWorkPath)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

		} else if strings.HasSuffix(file.Name(), ".mod") {
			err := addMod(filename)
			if err != nil {
				errs = multierror.Append(errs, err)
			} else {
				x.log("%s->added %s", prepend, filename)
			}
		}

	}

	return errs.ErrorOrNil()
}

func (x *handle) log(format string, a ...any) {
	if x.Debug {
		fmt.Fprint(os.Stderr, fmt.Sprintf(format, a...)+"\n")
	}
}
