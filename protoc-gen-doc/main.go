package main

import (
	"bytes"
	"code.witches.io/go/grpc-doc/proto"
	"fmt"
	proto2 "github.com/golang/protobuf/proto"
	"golang.org/x/net/html"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var n struct {
	Services, Messages, Enums map[string][]string
}

var typeToFile map[string]string

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

	typeToFile = make(map[string]string)
	n.Services = make(map[string][]string)
	n.Messages = make(map[string][]string)
	n.Enums = make(map[string][]string)

	var addTypes func(r *FileDescriptorProto, d *DescriptorProto, p, pr string)
	addTypes = func(r *FileDescriptorProto, d *DescriptorProto, p, pr string) {
		typeToFile[p+"."+d.GetName()] = r.GetName()
		n.Messages[r.GetName()] = append(n.Messages[r.GetName()], pr+d.GetName())
		for _, t := range d.NestedType {
			addTypes(r, t, p+"."+d.GetName(), pr+d.GetName()+".")
		}
		for _, e := range d.EnumType {
			typeToFile[p+"."+d.GetName()+"."+e.GetName()] = r.GetName()
			n.Enums[r.GetName()] = append(n.Enums[r.GetName()], pr+d.GetName()+"."+e.GetName())
		}
	}

	for _, r := range request.ProtoFile {
		for _, m := range r.MessageType {
			addTypes(r, m, r.GetPackage(), "")
		}
		for _, e := range r.EnumType {
			typeToFile[r.GetPackage()+"."+e.GetName()] = r.GetName()
			n.Enums[r.GetName()] = append(n.Enums[r.GetName()], e.GetName())
		}
		for _, s := range r.Service {
			typeToFile[r.GetPackage()+"."+s.GetName()] = r.GetName()
			n.Services[r.GetName()] = append(n.Services[r.GetName()], s.GetName())
		}
	}

	for _, r := range request.ProtoFile {
		dir, prof := filepath.Split(r.GetName())
		linkBase := strings.Repeat("../", len(strings.Split(dir, "/"))-1)

		g := generator{new(bytes.Buffer), typeToFile, linkBase, nil}

		rt := &generatorNode{
			Node: &html.Node{
				Type: html.DocumentNode,
			},
		}
		rt.AppendChild(&html.Node{
			Type: html.DoctypeNode,
			Data: "html",
		})

		ht := rt.E("html", html.Attribute{Key: "lang", Val: "ja"}).E("head")
		ht.E("meta", html.Attribute{Key: "charset", Val: "UTF-8"})
		ht.E("title").T(prof)
		ht.E("style").T(`
details > summary {
}

details[open] > summary {
}

li {
	list-style: none;
	white-space: nowrap;
}

li > .file {
	display: list-item;
}

.container {
	max-width: 60em;
	margin: 0 auto;
	font-family: monospace;
}

.comment {
	color: green;
	white-space: pre;
	display: block;
}

.comment, .struct {
	margin-top: 1em;
}

li:first-child > .comment,
tr:first-child > td > .comment,
struct:first-child,
.comment.trailing {
	margin-top: 0;
}

.keyword {
	color: maroon;
}

.type {
	color: darkorange;
}

.value {
	color: navy;
}

navbar {
	float: left;
	position: fixed;
}`)
		bt := ht.E("body")

		g.loadComments(r)
		g.generateNavigation(request.ProtoFile, r.GetName(), bt.E("navbar"))

		bt = bt.E("div", html.Attribute{Key: "class", Val: "container"})
		bt.E("h1").T(prof)

		ot := bt.E("ol")

		var ss []struct {
			Service *ServiceDescriptorProto
			Index   int
		}
		for i, s := range r.Service {
			ss = append(ss, struct {
				Service *ServiceDescriptorProto
				Index   int
			}{
				Service: s,
				Index:   i,
			})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Service.GetName() < ss[j].Service.GetName()
		})
		for _, s := range ss {
			g.generateService(s.Service, fmt.Sprintf("6,%d", s.Index), ot.E("li", html.Attribute{Key: "class", Val: "struct"}))
		}

		var es []struct {
			Enum  *EnumDescriptorProto
			Index int
		}
		for i, e := range r.EnumType {
			es = append(es, struct {
				Enum  *EnumDescriptorProto
				Index int
			}{
				Enum:  e,
				Index: i,
			})
		}
		sort.Slice(es, func(i, j int) bool {
			return es[i].Enum.GetName() < es[j].Enum.GetName()
		})
		for _, e := range es {
			g.generateEnum("", fmt.Sprintf("5,%d", e.Index), e.Enum, ot.E("li", html.Attribute{Key: "class", Val: "struct"}))
		}

		var ms []struct {
			Message *DescriptorProto
			Index   int
		}
		for i, m := range r.MessageType {
			ms = append(ms, struct {
				Message *DescriptorProto
				Index   int
			}{
				Message: m,
				Index:   i,
			})
		}
		sort.Slice(ms, func(i, j int) bool {
			return ms[i].Message.GetName() < ms[j].Message.GetName()
		})
		for _, m := range ms {
			g.generateMessage(r, "\t", fmt.Sprintf("4,%d", m.Index), m.Message, ot.E("li", html.Attribute{Key: "class", Val: "struct"}))
		}

		err := html.Render(g, rt.Node)
		if err != nil {
			panic(err)
		}

		file := &CodeGeneratorResponse_File{
			Name:    func(s string) *string { return &s }(strings.Replace(r.GetName(), ".proto", ".html", 1)),
			Content: func(s string) *string { return &s }(g.String()),
		}

		response.File = append(response.File, file)
	}

	response.File = append(response.File, index(request.ProtoFile))

	data, err = proto.Marshal(&response)
	if err != nil {
		panic(err)
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		panic(err)
	}
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

func index(rs []*FileDescriptorProto) *CodeGeneratorResponse_File {
	g := generator{new(bytes.Buffer), typeToFile, "", nil}

	rt := &generatorNode{
		Node: &html.Node{
			Type: html.DocumentNode,
		},
	}
	rt.AppendChild(&html.Node{
		Type: html.DoctypeNode,
		Data: "html",
	})

	ht := rt.E("html", html.Attribute{Key: "lang", Val: "ja"}).E("head")
	ht.E("meta", html.Attribute{Key: "charset", Val: "UTF-8"})
	ht.E("title").T("Protocol Buffers")
	ht.E("style").T(`
details > summary {
}

details[open] > summary {
}

li {
	list-style: none;
	white-space: nowrap;
}

li > .file {
	display: list-item;
}

.container {
	max-width: 60em;
	margin: 0 auto;
	font-family: monospace;
}

.comment {
	color: green;
	white-space: pre;
	display: block;
}

.comment, .struct {
	margin-top: 1em;
}

li:first-child > .comment,
tr:first-child > td > .comment,
struct:first-child,
.comment.trailing {
	margin-top: 0;
}

.keyword {
	color: maroon;
}

.type {
	color: darkorange;
}

.value {
	color: navy;
}

navbar {
	float: left;
	position: fixed;
}`)
	bt := ht.E("body").E("div", html.Attribute{Key: "class", Val: "container"})
	bt.E("h1").T("Protocol Buffers")
	bt.E("h2").T("Files")
	ot := bt.E("ol")
	for _, r := range rs {
		ot.E("li").E("a", html.Attribute{Key: "href", Val: strings.Replace(r.GetName(), ".proto", ".html", -1)}).T(r.GetName())
	}
	if len(n.Services) > 0 {
		bt.E("h2").T("Services")
		ot := bt.E("ol")
		var ps []string
		for p := range n.Services {
			ps = append(ps, p)
		}
		sort.Strings(ps)
		for _, p := range ps {
			for _, s := range n.Services[p] {
				g.addTypeLink(ot.E("li"), &FieldDescriptorProto{TypeName: proto.String("." + strings.Replace(filepath.Dir(p), "/", ".", -1) + "." + s)})
				ot.LastChild.LastChild.LastChild.LastChild.Data = filepath.Dir(p) + "." + s

			}
		}
	}
	if len(n.Enums) > 0 {
		bt.E("h2").T("Enums")
		ot := bt.E("ol")
		var ps []string
		for p := range n.Enums {
			ps = append(ps, p)
		}
		sort.Strings(ps)
		for _, p := range ps {
			for _, s := range n.Enums[p] {
				g.addTypeLink(ot.E("li"), &FieldDescriptorProto{TypeName: proto.String("." + strings.Replace(filepath.Dir(p), "/", ".", -1) + "." + s)})
				ot.LastChild.LastChild.LastChild.LastChild.Data = filepath.Dir(p) + "." + s
			}
		}
	}
	if len(n.Messages) > 0 {
		bt.E("h2").T("Messages")
		ot := bt.E("ol")
		var ps []string
		for p := range n.Messages {
			ps = append(ps, p)
		}
		sort.Strings(ps)
		for _, p := range ps {
			for _, s := range n.Messages[p] {
				g.addTypeLink(ot.E("li"), &FieldDescriptorProto{TypeName: proto.String("." + strings.Replace(filepath.Dir(p), "/", ".", -1) + "." + s)})
				ot.LastChild.LastChild.LastChild.LastChild.Data = filepath.Dir(p) + "." + s
			}
		}
	}

	err := html.Render(g, rt.Node)
	if err != nil {
		panic(err)
	}

	return &CodeGeneratorResponse_File{
		Name:    proto2.String("index.html"),
		Content: proto2.String(g.String()),
	}
}
