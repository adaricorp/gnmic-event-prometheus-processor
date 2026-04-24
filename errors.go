package main

import "errors"

var (
	ErrMissingSourceTag  = errors.New("source tag missing for event")
	ErrSourceNotCiscoWLC = errors.New("source not configured as cisco wlc")
	ErrTagExists         = errors.New("tag already exists")
	ErrEnumUnknownValue  = errors.New("unknown enum value")
)
