package core

import "github.com/richardlehane/siegfried/pkg/core/siegreader"

type Identifier interface {
	Identify(*siegreader.Buffer, chan Identification)
}

type Identification interface {
	String() string
	Confidence() float64 // how certain is this identification?
}
