// +build fpr,go1.4

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

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/cmd/sf/fpr"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var fprflag = flag.String("fpr", "false", "start siegfried fpr server e.g. -fpr localhost:5138")

type fprServer struct {
	*siegfried.Siegfried
}

func reply(p, e string) (*fpr.Reply, error) {
	return &fpr.Reply{
		p,
		e,
	}, nil
}

func (f *fprServer) Identify(ctx context.Context, in *fpr.Request) (*fpr.Reply, error) {
	fi, err := os.Open(in.Path)
	defer fi.Close()
	if err != nil {
		return reply("", "Error opening "+in.Path+"; error message: "+err.Error())
	}
	c, err := f.Siegfried.Identify(in.Path, fi)
	if err != nil {
		return reply("", "Error scanning "+in.Path+"; error message: "+err.Error())
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
		return reply("", "Error scanning "+in.Path+": no puids returned")
	case 1:
		if ids[0] == "UNKNOWN" {
			return reply("", "Unknown format: "+warn)
		}
		return reply(ids[0], "")
	default:
		return reply("", "Multiple formats returned: "+strings.Join(ids, ", "))
	}
}

func serveFpr(port string, s *siegfried.Siegfried) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error: failed to listen: %v", err)
	}
	g := grpc.NewServer()
	fpr.RegisterFprServer(g, &fprServer{s})
	g.Serve(lis)
}
