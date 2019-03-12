package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/html"
	"path/filepath"
	"sort"
	"strings"
)

func (g *generator) generateNavigation(fs []*FileDescriptorProto, open string, node *generatorNode) {
	type directory struct {
		directories map[string]*directory
		entries     map[string]struct{}
		open        bool
	}

	paths := &directory{
		directories: make(map[string]*directory),
		entries:     make(map[string]struct{}),
	}

	for _, f := range fs {
		path, file := filepath.Split(f.GetName())

		parent := paths
		directories := strings.Split(strings.TrimSuffix(path, "/"), "/")

		for len(directories) > 0 {
			entry, ok := parent.directories[directories[0]]
			if !ok {
				entry = &directory{
					directories: make(map[string]*directory),
					entries:     make(map[string]struct{}),
				}

				parent.directories[directories[0]] = entry
			}
			parent = entry
			directories = directories[1:]
		}

		parent.entries[file] = struct{}{}
	}

	parent := paths
	parent.open = true

	for _, d := range strings.Split(filepath.Dir(open), "/") {
		parent = parent.directories[d]
		parent.open = true
	}

	var attr = []html.Attribute{{Key: "open", Val: "open"}}

	var generate func(*directory, string, *generatorNode)
	generate = func(d *directory, path string, node *generatorNode) {
		ut := node.E("ul")
		var ds []string
		for p := range d.directories {
			ds = append(ds, p)
		}
		sort.Strings(ds)
		for _, p := range ds {
			dt := ut.E("li").E("details")
			if d.directories[p].open {
				dt.Attr = attr
			}
			dt.E("summary").T(p)
			generate(d.directories[p], filepath.Join(path, p), dt.E("div"))
		}
		var es []string
		for p := range d.entries {
			es = append(es, p)
		}
		sort.Strings(es)
		for _, p := range es {
			level := len(strings.Split(filepath.Dir(open), "/"))
			ut.E("li").E("span", html.Attribute{Key: "class", Val: "file"}).E("a", html.Attribute{
				Key: "href", Val: fmt.Sprint(strings.Repeat("../", level), path, "/", strings.Replace(p, ".proto", ".html", 1)),
			}).T(p)
		}
	}

	dt := node.E("div")
	dt.E("h2").T("Files")
	generate(paths, "", dt)
	dt = node.E("div")
	if len(n.Services[open]) > 0 {
		dt.E("h2").T("Services")
		sort.Strings(n.Services[open])
		ot := dt.E("ot")
		for _, s := range n.Services[open] {
			g.addTypeLink(ot.E("li"), &FieldDescriptorProto{TypeName: proto.String("." + strings.Replace(filepath.Dir(open), "/", ".", -1) + "." + s)})
			l := &ot.LastChild.LastChild.LastChild.LastChild.Data
			if strings.HasPrefix(*l, filepath.Base(filepath.Dir(open))+".") {
				*l = strings.Split(*l, ".")[1]
			}
		}
	}
	if len(n.Enums[open]) > 0 {
		dt.E("h2").T("Enums")
		sort.Strings(n.Enums[open])
		ot := dt.E("ot")
		for _, s := range n.Enums[open] {
			g.addTypeLink(ot.E("li"), &FieldDescriptorProto{TypeName: proto.String("." + strings.Replace(filepath.Dir(open), "/", ".", -1) + "." + s)})
			l := &ot.LastChild.LastChild.LastChild.LastChild.Data
			if strings.HasPrefix(*l, filepath.Base(filepath.Dir(open))+".") {
				*l = strings.Split(*l, ".")[1]
			}
		}
	}
	if len(n.Messages[open]) > 0 {
		dt.E("h2").T("Messages")
		sort.Strings(n.Messages[open])
		ot := dt.E("ot")
		for _, s := range n.Messages[open] {
			g.addTypeLink(ot.E("li"), &FieldDescriptorProto{TypeName: proto.String("." + strings.Replace(filepath.Dir(open), "/", ".", -1) + "." + s)})
			l := &ot.LastChild.LastChild.LastChild.LastChild.Data
			if strings.HasPrefix(*l, filepath.Base(filepath.Dir(open))+".") {
				*l = strings.Split(*l, ".")[1]
			}
		}
	}
}
