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
	"strings"
)

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
}

.container {
	max-width: 60em;
	margin: 0 auto;
}

.comment {
	color: green;
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

		ut := bt.E("ul")

		sort.Slice(r.Service, func(i, j int) bool {
			return r.Service[i].GetName() < r.Service[j].GetName()
		})
		for i, s := range r.Service {
			g.generateService(s, fmt.Sprintf("6,%d", i), ut)
		}

		ut = bt.E("ul")

		sort.Slice(r.MessageType, func(i, j int) bool {
			return r.MessageType[i].GetName() < r.MessageType[j].GetName()
		})
		for i, m := range r.MessageType {
			g.generateMessage(r, "\t", fmt.Sprintf("4,%d", i), m, ut)
		}

		ut = bt.E("ul")
		for _, l := range r.SourceCodeInfo.Location {
			ut.T(fmt.Sprintf("%v%s", l.Path, l))
		}

		bt.E("p").T(fmt.Sprintf("%#v\n\n%#v", r, typeToFile))

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
