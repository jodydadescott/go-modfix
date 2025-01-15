// Copyright (c) 2024 BlueRock Security, Inc.

// cSpell:ignore cloudvendor

package src

import (
	"fmt"
	"strings"
)

// ModFileState represents the state of a ModFile.
type ModFileState string

const (

	// ModFileStateUndefined is undefined
	ModFileStateUndefined ModFileState = ""

	// CloudVendorClean is Clean
	ModFileStateClean ModFileState = "clean"

	// ModFileStateDirty is Dirty
	ModFileStateDirty ModFileState = "dirty"
)

var modFileStateTypes = []ModFileState{ModFileStateClean, ModFileStateDirty}

// NewModFileState returns ModFileState from specified string. If string is not a valid type then an error will be returned.
func NewModFileState(s string) (ModFileState, error) {

	switch strings.ToLower(s) {

	case "":
		return ModFileStateUndefined, nil

	case string(ModFileStateClean):
		return ModFileStateClean, nil

	case string(ModFileStateDirty):
		return ModFileStateDirty, nil

	}

	return ModFileStateUndefined, fmt.Errorf("value %s is not valid, expecting one of %s", s, modFileStateTypes)
}

// Values returns all the possible values of type
func (ModFileState) Values() []ModFileState {
	return modFileStateTypes
}

// Pointer returns a pointer of the instance type.
func (x ModFileState) Pointer() *ModFileState {
	return &x
}

// ModFileStateValues returns all the possible values of type
func ModFileStateValues() []ModFileState {
	return modFileStateTypes
}

// ReportState represents the state of a ModFile.
type ReportState string

const (

	// ReportStateUndefined is undefined
	ReportStateUndefined ReportState = ""

	// ReportStateSuccess is Success
	ReportStateSuccess ReportState = "Success"

	// ReportStateFail is Fail
	ReportStateFail ReportState = "Fail"

	// ReportStatePartial is Partial (Success and Fail)
	ReportStatePartial ReportState = "Partial"
)

var reportStateTypes = []ReportState{ReportStateSuccess, ReportStateFail, ReportStatePartial}

// NewReportState returns ReportState from specified string. If string is not a valid type then an error will be returned.
func NewReportState(s string) (ReportState, error) {

	switch strings.ToLower(s) {

	case "":
		return ReportStateUndefined, nil

	case string(ReportStateSuccess):
		return ReportStateSuccess, nil

	case string(ReportStateFail):
		return ReportStateFail, nil

	case string(ReportStatePartial):
		return ReportStatePartial, nil

	}

	return ReportStateUndefined, fmt.Errorf("value %s is not valid, expecting one of %s", s, reportStateTypes)
}

// Values returns all the possible values of type
func (ReportState) Values() []ReportState {
	return reportStateTypes
}

// Pointer returns a pointer of the instance type.
func (x ReportState) Pointer() *ReportState {
	return &x
}

// ReportStateValues returns all the possible values of type
func ReportStateValues() []ReportState {
	return reportStateTypes
}
