package src

// cSpell:ignore WORKPATH MODFIX DRYRUN

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type Config struct {
	GoPath       string
	WorkPath     string
	Recursive    bool
	EnableHidden bool
	Debug        bool
	Verbose      bool
	DryRun       bool
	AddAll       bool
}

// LoadFromEnv loads values for this structs attributes if present from the shell environment.
// No checking is done to test if an attribute is already set hence this function should be called first.
func (x *Config) LoadFromEnv() error {

	var err error
	var errs *multierror.Error

	{
		v := os.Getenv("GOPATH")
		if v != "" {
			x.GoPath = v
		}
	}
	{
		v := os.Getenv("WORKPATH")
		if v != "" {
			x.WorkPath = v
		}
	}
	{
		v := os.Getenv("MODFIX_RECURSIVE")
		if v != "" {
			x.Recursive, err = strconv.ParseBool(v)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	{
		v := os.Getenv("MODFIX_HIDDEN")
		if v != "" {
			x.EnableHidden, err = strconv.ParseBool(v)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	{
		v := os.Getenv("MODFIX_DEBUG")
		if v != "" {
			x.Debug, err = strconv.ParseBool(v)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	{
		v := os.Getenv("MODFIX_VERBOSE")
		if v != "" {
			x.Verbose, err = strconv.ParseBool(v)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	{
		v := os.Getenv("MODFIX_DRYRUN")
		if v != "" {
			x.DryRun, err = strconv.ParseBool(v)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	{
		v := os.Getenv("MODFIX_ADD_ALL")
		if v != "" {
			x.AddAll, err = strconv.ParseBool(v)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return errs.ErrorOrNil()
}

type Report struct {
	WorkPathFiles []*ReportEntry
	GoPathFiles   []*ReportEntry
	State         ReportState
}

func (x *Report) Text() string {

	long := 0
	setLong := func(i []*ReportEntry) {
		for _, v := range i {
			if len(v.Filename) > long {
				long = len(v.Filename)
			}
			if len(v.Module) > long {
				long = len(v.Module)
			}
		}
	}
	setLong(x.WorkPathFiles)
	setLong(x.GoPathFiles)

	bar1 := strings.Repeat("#", long+15) + "\n"
	bar2 := strings.Repeat("-", long+15) + "\n"

	sb := &strings.Builder{}

	if x.GoPathFiles != nil {
		flag := false
		sb.WriteString(bar1)
		sb.WriteString("GoPathFiles:\n")
		sb.WriteString(bar1)
		for _, v := range x.GoPathFiles {
			if flag {
				sb.WriteString(bar2)
			} else {
				flag = true
			}
			v.text("  ", sb)
		}

		sb.WriteString(bar1)
		sb.WriteString("\n")
	}

	{
		flag := false
		sb.WriteString(bar1)
		sb.WriteString("WorkPathFiles:\n")
		sb.WriteString(bar1)

		for _, v := range x.WorkPathFiles {
			if flag {
				sb.WriteString(bar2)
			} else {
				flag = true
			}
			v.text("  ", sb)
		}

		sb.WriteString(bar1)
	}

	return sb.String()
}

type ReportEntry struct {
	Filename string
	Module   string
	State    *ModFileState
	Updated  *bool
	Error    error
}

func (x *ReportEntry) text(ident string, sb *strings.Builder) {

	sb.WriteString(fmt.Sprintf("%sFilename: %s\n", ident, x.Filename))
	sb.WriteString(fmt.Sprintf("%sModule: %s\n", ident, x.Module))

	if x.State != nil {
		sb.WriteString(fmt.Sprintf("%sState: %s\n", ident, *x.State))
	}

	if x.Updated != nil {
		sb.WriteString(fmt.Sprintf("%sUpdated: %t\n", ident, *x.Updated))
	}

	if x.Error != nil {
		sb.WriteString(fmt.Sprintf("%sError: %s\n", ident, x.Error))
	}
}
