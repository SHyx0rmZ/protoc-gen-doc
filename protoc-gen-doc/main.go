package main

import (
	"bytes"
	"code.witches.io/go/grpc-doc/proto"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type generator struct {
	*bytes.Buffer
	typeToFile map[string]string
	linkBase   string
	comments   map[string][]*SourceCodeInfo_Location
}

func (g *generator) P(args ...string) {
	for _, arg := range args {
		g.WriteString(arg)
	}
}

func main() {
	var request CodeGeneratorRequest

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	err = proto.Unmarshal(data, &request)
	if err != nil {
		panic(err)
	}

	response := CodeGeneratorResponse{
		File: make([]*CodeGeneratorResponse_File, 0),
	}

	typeToFile := make(map[string]string)

	for _, r := range request.ProtoFile {
		for _, m := range r.MessageType {
			typeToFile[r.GetPackage()+"."+m.GetName()] = r.GetName()
		}
	}

	for _, r := range request.ProtoFile {
		dir, prof := filepath.Split(r.GetName())
		linkBase := strings.Repeat("../", len(strings.Split(dir, "/"))-1)

		g := generator{new(bytes.Buffer), typeToFile, linkBase, nil}

		d := &html.Node{
			Type: html.DocumentNode,
		}
		d.AppendChild(&html.Node{
			Type: html.DoctypeNode,
			Data: "html",
		})
		nh := &html.Node{
			Type: html.ElementNode,
			Data: "html",
			Attr: []html.Attribute{
				{
					Key: "lang",
					Val: "ja",
				},
			},
		}
		d.AppendChild(nh)
		nh.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "head",
		})
		nh.LastChild.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "title",
		})
		nh.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: prof,
		})
		nh.LastChild.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "style",
		})
		nh.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: `
details > summary {
	list-style-image: url(/home/shyxormz/Desktop/未タイトルのフォルダ2/folder.svg);
}

details[open] > summary {
	list-style-image: url(/home/shyxormz/Desktop/未タイトルのフォルダ2/folder-open.svg);
}

li {
	list-style: none;
}

li > .file {
	display: list-item;
	list-style-image: url(/home/shyxormz/Desktop/未タイトルのフォルダ2/application-x-yaml.svg);
	margin-left: 80px;
}`,
		})
		nh.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "body",
		})
		nh.LastChild.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "h1",
		})
		nh.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: prof,
		})

		g.loadComments(r)
		g.generateNavigation(request.ProtoFile, r.GetName(), nh.LastChild)

		nh.LastChild.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "ul",
		})

		for i, m := range r.MessageType {
			g.generateMessage(r, "\t", fmt.Sprintf("4,%d", i), m, nh.LastChild.LastChild)
		}

		nh.LastChild.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "ul",
		})

		for i, s := range r.Service {
			g.generateService(s, fmt.Sprintf("6,%d", i), nh.LastChild.LastChild)
		}

		nh.LastChild.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "ul",
		})
		for _, l := range r.SourceCodeInfo.Location {
			nh.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: fmt.Sprintf("%v%s", l.Path, l),
			})
		}

		nh.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: fmt.Sprintf("%#v\n\n%#v", r, typeToFile),
		})

		err := html.Render(g, d)
		if err != nil {
			panic(err)
		}

		file := &CodeGeneratorResponse_File{
			Name:    func(s string) *string { return &s }(strings.Replace(r.GetName(), ".proto", ".html", 1)),
			Content: func(s string) *string { return &s }(g.String()),
		}

		response.File = append(response.File, file)
	}

	data, err = proto.Marshal(&response)
	if err != nil {
		panic(err)
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateService(s *ServiceDescriptorProto, path string, node *html.Node) {
	node.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "li",
	})
	node.LastChild.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "code",
	})
	node.LastChild.LastChild.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "pre",
	})
	node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: "service " + s.GetName() + " {\n",
	})

	for i, m := range s.GetMethod() {
		if i != 0 {
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "\n",
			})
		}
		loc, ok := g.comment(fmt.Sprintf("%s,2,%d", path, i))
		if ok {
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "\t//" + strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n\t//", -1) + "\n",
			})
		}
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: "\trpc " + m.GetName() + "(",
		})
		if m.GetClientStreaming() {
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "stream ",
			})
		}
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: m.GetInputType() + ") returns (",
		})
		if m.GetServerStreaming() {
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "stream ",
			})
		}
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: ")",
		})
		if m.Options != nil {
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: " {\n",
			})
			if m.Options.GetDeprecated() {
				node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
					Type: html.TextNode,
					Data: "\t\toption deprecated = true;\n",
				})
			}
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "\t}",
			})
		}
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: ";\n",
		})
	}
	node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: "}",
	})
}

var typeMap = map[FieldDescriptorProto_Type]string{
	FieldDescriptorProto_TYPE_DOUBLE:  "double",
	FieldDescriptorProto_TYPE_FLOAT:   "float",
	FieldDescriptorProto_TYPE_INT64:   "int64",
	FieldDescriptorProto_TYPE_UINT64:  "uint64",
	FieldDescriptorProto_TYPE_INT32:   "int32",
	FieldDescriptorProto_TYPE_FIXED64: "fixed64",
	FieldDescriptorProto_TYPE_FIXED32: "fixed32",
	FieldDescriptorProto_TYPE_BOOL:    "bool",
	FieldDescriptorProto_TYPE_STRING:  "string",
	// Tag-delimited aggregate.       :
	// Group type is deprecated and no:
	// implementations should still be:
	// treat group fields as unknown f:
	//FieldDescriptorProto_TYPE_GROUP   :
	//FieldDescriptorProto_TYPE_MESSAGE :
	// New in version 2.              :
	FieldDescriptorProto_TYPE_BYTES:  "bytes",
	FieldDescriptorProto_TYPE_UINT32: "uint32",
	//FieldDescriptorProto_TYPE_ENUM    : "enum",
	FieldDescriptorProto_TYPE_SFIXED32: "sfixed32",
	FieldDescriptorProto_TYPE_SFIXED64: "sfixed64",
	FieldDescriptorProto_TYPE_SINT32:   "sint32",
	FieldDescriptorProto_TYPE_SINT64:   "sint64",
}

func (g *generator) generateNavigation(fs []*FileDescriptorProto, open string, node *html.Node) {
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

	var generate func(*directory, string, *html.Node)
	generate = func(d *directory, path string, node *html.Node) {
		node.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "ul",
		})
		var ds []string
		for p := range d.directories {
			ds = append(ds, p)
		}
		sort.Strings(ds)
		for _, p := range ds {
			node.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "li",
			})
			node.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "details",
			})
			if d.directories[p].open {
				node.LastChild.LastChild.LastChild.Attr = []html.Attribute{{Key: "open", Val: "open"}}
			}
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "summary",
			})
			node.LastChild.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: p,
			})
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "div",
			})
			generate(d.directories[p], filepath.Join(path, p), node.LastChild.LastChild.LastChild.LastChild)
		}
		var es []string
		for p := range d.entries {
			es = append(es, p)
		}
		sort.Strings(es)
		for _, p := range es {
			node.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "li",
			})
			node.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "span",
				Attr: []html.Attribute{
					{
						Key: "class",
						Val: "file",
					},
				},
			})
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "a",
				Attr: []html.Attribute{
					{
						Key: "href",
						Val: strings.Repeat("../", len(strings.Split(filepath.Dir(open), "/"))) + path + "/" + strings.Replace(p, ".proto", ".html", 1),
					},
				},
			})
			node.LastChild.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: p,
			})
		}
	}

	generate(paths, "", node)
}

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

func (g *generator) generateMessage(f *FileDescriptorProto, indent, path string, d *DescriptorProto, node *html.Node) {
	indent = "\t"

	node.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "li",
		Attr: []html.Attribute{
			{
				Key: "id",
				Val: d.GetName(),
			},
		},
	})
	node.LastChild.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "code",
	})
	node.LastChild.LastChild.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "pre",
	})

	loc, ok := g.comment(path)
	if ok {
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: "//" + strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n//", -1) + "\n",
		})
	}
	node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: "message " + d.GetName() + " {\n",
	})

	if len(d.NestedType) > 0 {
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: "ul",
		})
		for i, message := range d.NestedType {
			g.generateMessage(f, indent+"\t", fmt.Sprintf("%s,3,%d", path, i), message, node.LastChild.LastChild.LastChild.LastChild)
		}
	}

	if len(d.GetReservedRange()) > 0 {
		var rs []string
		for _, x := range d.GetReservedRange() {
			if *x.Start != *x.End-1 {
				rs = append(rs, fmt.Sprintf("%d to %d", *x.Start, *x.End-1))
			} else {
				rs = append(rs, strconv.Itoa(int(*x.Start)))
			}
		}
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: indent + "reserved " + strings.Join(rs, ", ") + ";\n",
		})
	}

	if len(d.GetReservedName()) > 0 {
		var rs []string
		for _, x := range d.GetReservedName() {
			rs = append(rs, `"`+x+`"`)
		}
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: indent + "reserved " + strings.Join(rs, ", ") + ";\n",
		})
	}

	for i, field := range d.Field {

		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: indent,
		})

		loc, ok := g.comment(fmt.Sprintf("%s,2,%d", path, i))
		if ok {
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "//" + strings.Replace(strings.Trim(*loc.LeadingComments, "\n"), "\n", "\n"+indent+"//", -1) + "\n" + indent,
			})

		}

		tn := field.GetTypeName()
		tnp := strings.Split(tn, ".")[1:]
		if len(tnp) > 1 {
			o := len(tnp) - 2
			if o < 0 {
				o = 0
			}
			tn = strings.Join(tnp[o:], ".")
			//tl := strings.Replace(strings.Join(tnp[:len(tnp)-1], "/"), ".proto", ".html", 1) + "#" + tnp[len(tnp)-1]
			tl := strings.Replace(g.typeToFile[field.GetTypeName()[1:]], ".proto", ".html", 1) + "#" + tnp[len(tnp)-1]

			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.ElementNode,
				Data: "a",
				Attr: []html.Attribute{
					{
						Key: "href",
						Val: g.linkBase + tl,
					},
				},
			})
			node.LastChild.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: tn,
			})
		} else {
			if len(tn) == 0 {
				node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
					Type: html.TextNode,
					Data: typeMap[field.GetType()],
				})
			} else {
				node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
					Type: html.TextNode,
					Data: tn,
				})
			}
		}

		//if len(field.GetTypeName()) == 0 {
		//	g.P(fmt.Sprintf("%T %#v", field.GetType(), field.GetType()))
		//}
		//
		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: " " + field.GetName() + " = " + strconv.Itoa(int(field.GetNumber())),
		})

		if field.Options != nil {
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: " [\n",
			})
			if field.Options.GetDeprecated() {
				node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
					Type: html.TextNode,
					Data: indent + "\tdeprecated = true\n",
				})
			}
			node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: indent + "];",
			})
		}

		node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: "\n\n",
		})
	}
	node.LastChild.LastChild.LastChild.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: indent[:len(indent)-1] + "}",
	})
}

//func (g *generator)
