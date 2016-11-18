// +build !windows

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
	"github.com/richardlehane/siegfried/pkg/config"
)

var fprflag = flag.Bool("fpr", false, "start siegfried fpr server at "+config.Fpr())

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
		return reply("error: failed to open " + path + "; got " + err.Error())
	}
	ids, err := s.Identify(fi, path, "")
	if ids == nil {
		return reply("error: failed to scan " + path + "; got " + err.Error())
	}
	switch len(ids) {
	case 0:
		return reply("error: scanning " + path + ": no formats returned")
	case 1:
		if !ids[0].Known() {
			return reply("error: format unknown; got " + ids[0].Warn())
		}
		return reply(ids[0].String())
	default:
		strs := make([]string, len(ids))
		for i, v := range ids {
			strs[i] = v.String()
		}
		return reply("error: multiple formats returned; got " + strings.Join(strs, ", "))
	}
}

func serveFpr(addr string, s *siegfried.Siegfried) {
	// remove the socket file if it exists
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
	buf := make([]byte, 4024)
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("FPR error: bad connection: %v", err)
		}
		l, err := conn.Read(buf)
		if err != nil {
			conn.Write([]byte("error reading from connection: " + err.Error()))
		} else {
			conn.Write(fpridentify(s, string(buf[:l])))
		}
		conn.Close()
	}
}
