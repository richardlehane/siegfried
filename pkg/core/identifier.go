package core

import (
	"sync"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Identifier interface {
	Identify(*siegreader.Buffer, string, chan Identification, *sync.WaitGroup)
}

type Identification interface {
	String() string
	Confidence() float64 // how certain is this identification?
}
