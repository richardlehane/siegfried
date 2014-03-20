package pronom

import "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"

// This file a work in progress...

type Format struct {
}

func (f *Format) Signatures() []frames.Signature {
	return nil
}

/*
func (f *Format) Prefer() []core.Format { return nil }

func (f *Format) Super() []core.Format { return nil }
*/
func (f *Format) String(filename string) string { return "" }

func (f *Format) JSON(filename string) string { return "" }

func (f *Format) XML(filename string) string { return "" }

func (f *Format) CSV(filename string) string { return "" }
