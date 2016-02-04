A reader for the [WARC and ARC web archive formats](http://iipc.github.io/warc-specifications/).

**Note**: This package has been written for use in [https://github.com/richardlehane/siegfried](https://github.com/richardlehane/siegfried) and has a bunch of quirks relating to that use case. If you're after a general purpose golang WARC package, you might be better suited by one of these excellent choices:

  - [https://github.com/edsu/warc](https://github.com/edsu/warc)
  - [https://github.com/slyrz/warc](https://github.com/slyrz/warc)

Example usage:

```go
f, _ := os.Open("examples/IAH-20080430204825-00000-blackbook.arc")
// NewReader(io.Reader) can be used to read WARC, ARC or gzipped WARC or ARC files
rdr, err := webarchive.NewReader(f)
if err != nil {
  log.Fatal(err)
}
// use Next() to iterate through all records in the WARC or ARC file
for record, err := rdr.Next(); err == nil; record, err = rdr.Next() {
  // records implement the io.Reader interface
  i, err := io.Copy(ioutil.Discard, record)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Read: %d bytes\n", i)
  // records also have URL(), Date() and Size() methods
  fmt.Printf("URL: %s, Date: %v, Size: %d\n", record.URL(), record.Date(), record.Size())
  // the Fields() method returns all the fields in the WARC or ARC record
  for key, values := range record.Fields() {
    fmt.Printf("Field key: %s, Field values: %v\n", key, values)
  }
}
f, _ = os.Open("examplesIAH-20080430204825-00000-blackbook.warc.gz")
defer f.Close()
// readers can Reset() to reuse the underlying buffers
err = rdr.Reset(f)
// the Close() method should be used if you pass in gzipped files, it is a nop for 
// non-gzipped files
defer rdr.Close()
// NextPayload() skips non-resource, conversion or response records and merges 
// continuations into single records. It also strips HTTP headers from response 
// records. After stripping, those HTTP headers are available alongside the WARC 
// headers in the record.Fields() map.
for record, err := rdr.NextPayload(); err == nil; record, err = rdr.NextPayload() {
  i, err := io.Copy(ioutil.Discard, record)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Read: %d bytes\n", i)
  // any skipped HTTP headers can be retrieved from the Fields() map
  for key, values := range record.Fields() {
    fmt.Printf("Field key: %s, Field values: %v\n", key, values)
  }
}
```
  
Install with `go get github.com/richardlehane/webarchive`

[![Build Status](https://travis-ci.org/richardlehane/webarchive.png?branch=master)](https://travis-ci.org/richardlehane/webarchive) [![GoDoc](https://godoc.org/github.com/richardlehane/webarchive?status.svg)](https://godoc.org/github.com/richardlehane/webarchive)
