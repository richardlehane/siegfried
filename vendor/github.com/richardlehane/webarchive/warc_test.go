package webarchive

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

func TestWARC(t *testing.T) {
	f, _ := os.Open("examples/hello-world.warc")
	defer f.Close()
	rdr, err := NewWARCReader(f)
	if err != nil {
		t.Fatal("failure loading example: " + err.Error())
	}
	rec, err := rdr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if rec.Date().Format(time.RFC3339) != "2015-07-08T21:55:13Z" {
		t.Errorf("expecting 2015-07-08T21:55:13Z, got %v", rec.Date())
	}
}

func TestGZ(t *testing.T) {
	f, _ := os.Open("examples/IAH-20080430204825-00000-blackbook.warc.gz")
	defer f.Close()
	rdr, err := NewWARCReader(f)
	if err != nil {
		t.Fatal("failure loading example: " + err.Error())
	}
	defer rdr.Close()
	var count int
	for _, err = rdr.NextPayload(); err != io.EOF; _, err = rdr.NextPayload() {
		if err != nil {
			log.Fatal(err)
		}
		count++
	}
	if count != 299 {
		t.Errorf("expecting 299 payloads, got %d", count)
	}
}

func ExampleNewWARCReader() {
	f, _ := os.Open("examples/IAH-20080430204825-00000-blackbook.warc")
	rdr, err := NewWARCReader(f)
	if err != nil {
		log.Fatal("failure creating an warc reader")
	}
	rec, err := rdr.NextPayload()
	if err != nil {
		log.Fatal("failure seeking: " + err.Error())
	}
	buf := make([]byte, 55)
	io.ReadFull(rec, buf)
	var count int
	wrec, ok := rec.(WARCRecord)
	if !ok {
		log.Fatal("failure doing WARCRecord interface assertion")
	}
	fmt.Println(wrec.ID())
	for _, err = rdr.NextPayload(); err != io.EOF; _, err = rdr.NextPayload() {
		if err != nil {
			log.Fatal(err)
		}
		count++
	}
	fmt.Printf("%s\n%d", buf, count)
	// Output:
	// <urn:uuid:ff728363-2d5f-4f5f-b832-9552de1a6037>
	// 20080430204825
	// www.archive.org.	589	IN	A	207.241.229.39
	// 298
}
