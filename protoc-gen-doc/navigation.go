package main

import (
	"fmt"
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
			ut.E("li").E("span").E("a", html.Attribute{
				Key: "href", Val: fmt.Sprint(strings.Repeat("../", level), path, "/", strings.Replace(p, ".proto", ".html", 1)),
			}).T(p)
		}
	}

	generate(paths, "", node)
}
