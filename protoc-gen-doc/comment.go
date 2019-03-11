package main

import (
	"strconv"
	"strings"
)

func (g *generator) loadComments(f *FileDescriptorProto) {
	g.comments = make(map[string][]*SourceCodeInfo_Location)
	for _, loc := range f.SourceCodeInfo.Location {
		if loc.LeadingComments == nil {
			continue
		}
		var p []string
		for _, n := range loc.Path {
			p = append(p, strconv.Itoa(int(n)))
		}
		path := strings.Join(p, ",")
		//g.P("- ", path, loc.String(), "\n")
		g.comments[path] = append(g.comments[path], loc)
	}
}

func (g *generator) comment(path string) (*SourceCodeInfo_Location, bool) {
	locs, ok := g.comments[path]
	if !ok {
		return nil, false
	}
	return locs[0], true
}
