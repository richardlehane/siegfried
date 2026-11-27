//go:build js

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"hash"
	"io"
	"log"
	"path/filepath"
	"strings"
	"syscall/js"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/internal/checksum"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/decompress"
	"github.com/richardlehane/siegfried/pkg/pronom"
	"github.com/richardlehane/siegfried/pkg/writer"
)

type output int

const (
	jsonOut output = iota
	yamlOut
	csvOut
	droidOut
)

//go:embed signature_files/droid_signature_file.xml
var droid string

//go:embed  signature_files/droid_container_signature_file.xml
var container string

const signatureDetails string = "WASM Embedded Defaults"
const signatureFileVersion string = "v125"
const containerFileVersion string = "2026-10-27"

var sf *siegfried.Siegfried

func opts(args []js.Value) (output, checksum.HashTyp, bool) {
	var out output
	var ht checksum.HashTyp = -1
	var z bool
	for _, v := range args {
		vv := v.String()
		switch vv {
		case "yaml":
			out = yamlOut
		case "csv":
			out = csvOut
		case "droid":
			out = droidOut
		case "z":
			z = true
			config.SetArchiveFilterPermissive(config.ListAllArcTypes())
		default:
			htyp := checksum.GetHash(v.String())
			if htyp > -1 {
				ht = htyp
			}
		}
	}
	return out, ht, z
}

func identifyRdr(
	s *siegfried.Siegfried,
	r io.Reader,
	w writer.Writer,
	path string,
	mime string,
	sz int64,
	mod time.Time,
	h hash.Hash,
	z bool,
	do bool,
) {
	b, berr := s.Buffer(r)
	defer s.Put(b)
	ids, err := s.IdentifyBuffer(b, berr, path, mime)
	if ids == nil {
		w.File(path, sz, mod.Format(time.RFC3339), nil, err, ids)
		return
	}
	// calculate checksum
	var cs []byte
	if h != nil {
		var i int64
		l := h.BlockSize()
		for ; ; i += int64(l) {
			buf, _ := b.Slice(i, l)
			if buf == nil {
				break
			}
			h.Write(buf)
		}
		cs = h.Sum(nil)
	}
	// decompress if an archive format
	if !z {
		w.File(path, sz, mod.Format(time.RFC3339), cs, err, ids)
		return
	}
	arc := decompress.IsArc(ids)
	if arc == config.None {
		w.File(path, sz, mod.Format(time.RFC3339), cs, err, ids)
		return
	}
	d, err := decompress.New(arc, b, path)
	if err != nil {
		w.File(path, sz, mod.Format(time.RFC3339), cs, fmt.Errorf("failed to decompress, got: %v", err), ids)
		return
	}
	// write the result
	w.File(path, sz, mod.Format(time.RFC3339), cs, err, ids)
	// decompress and recurse
	for err = d.Next(); ; err = d.Next() {
		if err != nil {
			if err == io.EOF {
				return
			}
			w.File(decompress.Arcpath(path, ""), 0, time.Time{}.Format(time.RFC3339), nil, fmt.Errorf("error occurred during decompression: %v", err), nil)
			return
		}
		if do {
			for _, v := range d.Dirs() {
				w.File(v, -1, "", nil, nil, nil)
			}
		}
		identifyRdr(s, d.Reader(), w, d.Path(), d.MIME(), d.Size(), d.Mod(), h, z, do)
	}
}

func identifyFile(
	s *siegfried.Siegfried,
	r *reader,
	f js.Value,
	w writer.Writer,
	h hash.Hash,
	z bool,
	do bool,
	dirs []string,
) error {
	var (
		name string
		mod  time.Time
	)
	promise := f.Call("getFile")
	val, err := await(promise)
	if err != nil {
		return err
	}
	name = filepath.Join(append(dirs, val.Get("name").String())...)
	modUnix := int64(val.Get("lastModified").Int())
	mod = time.UnixMilli(modUnix)
	r.reset(val)
	identifyRdr(s, r, w, name, "", r.Size(), mod, h, z, do)
	return nil
}

func identifyFiles(
	s *siegfried.Siegfried,
	r *reader,
	fsh js.Value,
	w writer.Writer,
	h hash.Hash,
	z bool,
	do bool,
	dirs []string,
) error {
	kind := fsh.Get("kind").String()
	if kind == "file" {
		return identifyFile(s, r, fsh, w, h, z, do, dirs)
	}
	dirs = append(dirs, fsh.Get("name").String())
	entries := fsh.Call("values")
	for {
		next := entries.Call("next")
		entry, err := await(next)
		if err != nil {
			return err
		}
		if entry.Get("done").Bool() {
			return nil
		}
		fsh := entry.Get("value")
		err = identifyFiles(s, r, fsh, w, h, z, do, dirs)
		if err != nil {
			return err
		}
	}
}

// identify(FileSystemHandle, ...OPTS)
// OPTS: json (default), csv, yaml, droid,
//
//	     md5, sha1, sha256, sha512, crc,
//		 z
func sfWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			panic("SF WASM error: provide a FileSystemHandle as first argument")
		}
		promiseHandler := js.FuncOf(func(v js.Value, x []js.Value) interface{} {
			resolve := x[0]
			reject := x[1]
			go func() {
				o, ht, z := opts(args[1:])
				h := checksum.MakeHash(ht)
				out := &bytes.Buffer{}
				var w writer.Writer
				r := newReader()
				switch o {
				case yamlOut:
					w = writer.YAML(out)
				case csvOut:
					w = writer.CSV(out)
				case droidOut:
					w = writer.Droid(out)
				default:
					w = writer.JSON(out)
				}
				w.Head(config.SignatureBase(), time.Now(), sf.C, config.Version(), sf.Identifiers(), sf.Fields(), ht.String())
				fsh := args[0]
				err := identifyFiles(sf, r, fsh, w, h, z, o == droidOut, nil)
				w.Tail()
				if err != nil {
					reject.Invoke(err.Error())
				} else {
					resolve.Invoke(out.String())
				}
			}()
			return nil
		})
		// Create and return the Promise object
		promise := js.Global().Get("Promise")
		return promise.New(promiseHandler)
	})
}

type extension struct {
	extData []byte
	fname   string
}

func makeSig(extension []extension) (*siegfried.Siegfried, error) {

	fname := fmt.Sprintf("temp filename: %d", len(extension))

	opts := []config.Option{}
	if fname != "" {
		opts = append(opts,
			config.SetDetails(
				fmt.Sprintf("%s; %s; Container: %s; Extension: %s",
					signatureDetails,
					signatureFileVersion,
					containerFileVersion,
					fname,
				),
			),
		)
	} else {
		opts = append(opts,
			config.SetDetails(
				fmt.Sprintf("%s; %s; Container: %s",
					signatureDetails,
					signatureFileVersion,
					containerFileVersion,
				),
			),
		)
	}

	exts := []pronom.Exts{}

	for _, ext := range extension {
		newExt := pronom.Exts{}
		newExt.ExtData = ext.extData
		if strings.Contains(string(ext.extData), "<ContainerSignatureMapping") {
			newExt.Container = true
		}
		exts = append(exts, newExt)

	}

	id, err := pronom.NewFromByteArray([]byte(container), []byte(droid), exts, opts...)
	if err != nil {
		return nil, err
	}
	s := siegfried.New()
	err = s.Add(id)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func royWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		promiseHandler := js.FuncOf(func(v js.Value, x []js.Value) interface{} {
			// x is a slice containing two functions resolve() and reject()
			// which are called below when providing a response to the webapp.
			//
			// v is not invoked by royWrapper.
			//
			if len(args) < 1 {
				panic("provide a FileSystemHandle as first argument")
			}
			resolve := x[0]
			reject := x[1]
			go func() {
				out := &bytes.Buffer{}
				out.Write([]byte("Successfully loaded new signature file")) // written to the resolve function...
				var err error
				// Even pluralized, use the first arg as a canary.
				fshx := args[0]
				kind := fshx.Get("kind").String()
				if kind != "file" {
					panic("must provide a file to extend Siegfried's signatures")
				}
				exts := []extension{}
				for _, val := range args {
					if val.Type() == js.TypeString || val.Type() == js.TypeUndefined {
						// else... we have an object, which we assume to be a file.
						// this assumption might be challenged...
						continue
					}
					path := []string{}
					name := filepath.Join(append(path, val.Get("name").String())...)
					log.Println("signature file:", name)
					r := newReader()
					promise := val.Call("getFile")
					val, err := await(promise)
					if err != nil {
						panic("unable to read signature file to extend Siegfried's signatures")
					}
					r.reset(val)
					log.Println("extension signature file size:", len(r.buf))
					extensionx, _ := r.Slice(0, int(r.sz))
					ext := new(extension)
					ext.extData = extensionx
					ext.fname = name
					exts = append(exts, *ext)
				}
				sf, err = makeSig(exts)
				if err != nil {
					reject.Invoke(err.Error())
				} else {
					resolve.Invoke(out.String())
				}
			}()
			return nil
		})
		// Create and return the Promise object
		promise := js.Global().Get("Promise")
		return promise.New(promiseHandler)
	})
}

func main() {
	x := make([]extension, 2)
	sf, _ = makeSig(x)
	js.Global().Set("identify", sfWrapper())
	js.Global().Set("sigload", royWrapper())
	<-make(chan bool)
}
