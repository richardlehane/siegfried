// +build archivematica,linux

// Copyright 2015 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var fprflag = flag.String("fpr", "false", "start siegfried fpr server e.g. -fpr /tmp/siegfried/fpr")

func reply(s string) []byte {
	if len(s) > 1024 {
		return []byte(s[:1024])
	}
	return []byte(s)
}

func fpridentify(s *siegfried.Siegfried, path string) []byte {
	fi, err := os.Open(path)
	defer fi.Close()
	if err != nil {
		return reply("Error: failed to open " + path + "; error message: " + err.Error())
	}
	c, err := s.Identify(path, fi)
	if err != nil {
		return reply("Error: failed to scan " + path + "; error message: " + err.Error())
	}
	var ids []string
	var warn string
	for i := range c {
		ids = append(ids, i.String())
		if len(ids) == 1 && ids[0] == "UNKNOWN" {
			warn = i.(*pronom.Identification).Warning
		}
	}
	switch len(ids) {
	case 0:
		return reply("Error: scanning " + path + ": no puids returned")
	case 1:
		if ids[0] == "UNKNOWN" {
			return reply("Error: format unknown; message: " + warn)
		}
		return reply(ids[0])
	default:
		return reply("Error: multiple formats returned; message : " + strings.Join(ids, ", "))
	}
}

func serveFpr(addr string, s *siegfried.Siegfried) {
	// remove the address if it exists
	if _, err := os.Stat(addr); err == nil {
		os.Remove(addr)
	}
	uaddr, err := net.ResolveUnixAddr("unix", addr)
	if err != nil {
		log.Fatalf("FPR error: failed to get address: %v", err)
	}
	lis, err := net.ListenUnix("unix", uaddr)
	if err != nil {
		log.Fatalf("FPR error: failed to listen: %v", err)
	}
	defer os.Remove(addr)
	buf := make([]byte, 4024)
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("FPR error: bad connection: %v", err)
		}
		l, err := conn.Read(buf)
		if err != nil {
			conn.Write([]byte("Error: " + err.Error()))
		}
		conn.Write(fpridentify(s, string(buf[:l])))
		conn.Close()
	}
}
