# Siegfried WASM

## API

`identify(FileSystemHandle, ...options)` returns `Promise(String)`

The [FileSystemHandle](https://developer.mozilla.org/en-US/docs/Web/API/FileSystemHandle) can be either a file or directory handle. If a directory handle is given, siegfried will recurse all contents of that directory (files and subdirectories).

Options are the following strings, in any order:

  - "yaml", "csv", or "droid" to change output format
  - "md5", "sha1", "sha256", "sha512" or "crc" for a checksum
  - "z" to decompress archive formats.

The return value is a [Promise](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise) that resolves to an output string (default JSON).

## Install

To include in your own web page include the `wasm_exec.js` and `sf.wasm` files in your project and use the following script to load: 

    <script src="wasm_exec.js"></script>
    <script>
        const go = new window.Go();

        WebAssembly.instantiateStreaming(
            fetch("sf.wasm"),
            go.importObject
        ).then(
            (obj) => {
                go.run(obj.instance);
            }
        );
    </script>

Once loaded, the `identify` method is available to use. See the example (`example.js`) in this package.

## Building

To build the wasm file yourself do:

`GOOS=js GOARCH=wasm go build -o sf.wasm github.com/richardlehane/siegfried/wasm`

Signatures are embedded in the `sf.wasm` file using the [static package](https://github.com/richardlehane/siegfried/tree/main/pkg/static). You can customise the embedded signatures by editing the `gen.go` file in that package (edit the path in the first line of the main function) and then running `go generate`.

## Running the example

The example in this package won't run locally in a browser just by opening the "index.html" file. It needs to be accessed via a server. Any server capable of serving a file directory locally will work. If you don't have one installed, this simple go script works:

    package main

    import (
        "flag"
        "log"
        "net/http"
    )

    func main() {
        port := flag.String("p", "8100", "port to serve on")
        directory := flag.String("d", ".", "the directory of static file to host")
        flag.Parse()
        http.Handle("/", http.FileServer(http.Dir(*directory)))
        log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
        log.Fatal(http.ListenAndServe(":"+*port, nil))
    }

