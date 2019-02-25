package main

import (
	"bytes"
	"code.witches.io/go/grpc-doc/proto"
	"fmt"
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

		g := generator{new(bytes.Buffer), typeToFile, linkBase}

		g.WriteString(`<!DOCTYPE html><html lang="en"><head><title>`)
		g.WriteString(prof)
		g.WriteString(`</title>
<style>
details > summary {
	list-style-image: url(/home/shyxormz/Desktop/未タイトルのフォルダ2/folder.svg) 0;
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
</style>
<body><h1>`)
		g.WriteString(prof)
		g.WriteString(`</h1>`)
		g.generateNavigation(request.ProtoFile, r.GetName())

		g.WriteString(`<ul>`)

		for i, m := range r.MessageType {
			g.generateMessage(r, "\t", fmt.Sprintf("4,%d", i), m)
		}

		g.P("</ul><ul>")

		for _, s := range r.Service {
			g.P("<li><code>service ", s.GetName(), "</code></li>")
		}

		g.WriteString(`</ul>`)

		g.P("<ul>")
		for _, l := range r.SourceCodeInfo.Location {
			g.P("<li>", fmt.Sprintf("%v", l.Path))
			g.P(l.String())
			g.P("</li>")
		}
		g.P("</ul>")

		g.WriteString(`<code>`)
		g.WriteString(fmt.Sprintf("%#v\n\n%#v", r, typeToFile))
		g.WriteString(`</code></body></html>`)

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

func (g *generator) generateNavigation(fs []*FileDescriptorProto, open string) {
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

	var generate func(*directory, string)
	generate = func(d *directory, path string) {
		g.P(`<ul>`)
		var ds []string
		for p := range d.directories {
			ds = append(ds, p)
		}
		sort.Strings(ds)
		for _, p := range ds {
			if d.directories[p].open {
				g.P(`<li><details open><summary>`, p, `</summary><div>`)
			} else {
				g.P(`<li><details><summary>`, p, `</summary><div>`)
			}
			generate(d.directories[p], filepath.Join(path, p))
			g.P(`</div></li>`)
		}
		var es []string
		for p := range d.entries {
			es = append(es, p)
		}
		sort.Strings(es)
		for _, p := range es {
			g.P(`<li><span class="file"><a href="`, strings.Repeat("../", len(strings.Split(filepath.Dir(open), "/")))+path+"/"+strings.Replace(p, ".proto", ".html", 1), `">`, p, `</a></span></li>`)
		}
		g.P(`</ul>`)
	}

	generate(paths, "")
	//g.P(fmt.Sprintf("%#v", paths))
}

func (g *generator) generateMessage(f *FileDescriptorProto, indent, path string, d *DescriptorProto) {
	indent = "\t"

	comments := make(map[string]*SourceCodeInfo_Location)
	for _, loc := range f.SourceCodeInfo.Location {
		if loc.LeadingComments == nil {
			continue
		}
		var p []string
		for _, n := range loc.Path {
			p = append(p, strconv.Itoa(int(n)))
		}
		g.P("- ", strings.Join(p, ","), loc.String(), "\n")
		comments[strings.Join(p, ",")] = loc
	}

	g.P(`<li id="`, d.GetName(), `"><code><pre>`)

	loc, ok := comments[path]
	if ok {
		g.P("//", strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n"+"//", -1), "\n")
	}
	g.P(`message `, d.GetName(), " {\n")

	if len(d.NestedType) > 0 {
		g.P("<ul>")
		for i, message := range d.NestedType {
			g.generateMessage(f, indent+"\t", fmt.Sprintf("%s,3,%d", path, i), message)
		}
		g.P("</ul>")
	}

	if len(d.GetReservedRange()) > 0 {
		var rs []string
		for i, x := range d.GetReservedRange() {
			g.P("<", strconv.Itoa(i), ">")
			if *x.Start != *x.End-1 {
				rs = append(rs, fmt.Sprintf("%d to %d", *x.Start, *x.End-1))
			} else {
				rs = append(rs, strconv.Itoa(int(*x.Start)))
			}
		}
		g.P(indent, "reserved ", strings.Join(rs, ", "), ";\n")
	}

	if len(d.GetReservedName()) > 0 {
		var rs []string
		for _, x := range d.GetReservedName() {
			rs = append(rs, `"`+x+`"`)
		}
		g.P(indent, "reserved ", strings.Join(rs, ", "), ";\n")
	}

	for i, field := range d.Field {

		g.P(indent)

		//g.WriteString()

		loc, ok := comments[fmt.Sprintf("%s,2,%d", path, i)]
		if ok {
			g.P("//", strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n"+indent+"//", -1), "\n", indent)
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

			g.P(`<a href="`, g.linkBase, tl, `">`, tn, "</a>")
		} else {
			if len(tn) == 0 {
				g.P(typeMap[field.GetType()])
			} else {
				g.P(tn)
			}
		}

		//if len(field.GetTypeName()) == 0 {
		//	g.P(fmt.Sprintf("%T %#v", field.GetType(), field.GetType()))
		//}
		//
		g.P(" ", field.GetName(), " = ", strconv.Itoa(int(field.GetNumber())))

		if field.Options != nil {
			g.P(" [\n")
			if field.Options.GetDeprecated() {
				g.P(indent, "\tdeprecated = true\n")
			}
			g.P(indent, "];")
		}

		g.P("\n\n")
	}
	g.P(indent[:len(indent)-1], "}</pre></code></li>")
}

//func (g *generator)
