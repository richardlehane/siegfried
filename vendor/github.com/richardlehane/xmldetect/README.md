A library to detect whether the first and last bytes of a stream are valid xml and to report opening and closing tags, default namespaces, and the presence of an xml declaration.

Example usage:

```go
var ex = `
  <?xml version="1.0"?>
  <!DOCTYPE doc [
  <!ELEMENT doc (#PCDATA)>
  <!ENTITY e "<![CDATA[&foo;]]]]]]]]>">
  ]>
  <!-- ignore me -->
  <doc xmlns="ex">&e;</doc>
`
dec, tag, ns, err := xmldetect.Root(strings.NewReader(ex))
if err == nil && dec && tag == "doc" && ns == "ex" {
  fmt.Println("root tag is 'doc' with namespace 'ex'")
}
```
  
Install with `go get github.com/richardlehane/xmldetect`

[![Build Status](https://travis-ci.org/richardlehane/xmldetect.png?branch=master)](https://travis-ci.org/richardlehane/xmldetect) [![GoDoc](https://godoc.org/github.com/richardlehane/xmldetect?status.svg)](https://godoc.org/github.com/richardlehane/xmldetect)

Licensed under the Apache License, Version 2.0
