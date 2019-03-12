package main

import (
	"fmt"
	"golang.org/x/net/html"
	"sort"
	"strconv"
	"strings"
)

type x struct {
	Data  interface{}
	Path  string
	Type  sortType
	Field *int32
	Name  string
}

type sortType int

const (
	sortTypeNumber sortType = iota
	sortTypeReserved
	sortTypeEnum
	sortTypeMessage
)

type oow struct {
	Field *FieldDescriptorProto
	Oneof *OneofDescriptorProto
}

func (o oow) String() string { return o.Field.String() }

func (g *generator) generateMessage(f *FileDescriptorProto, indent, path string, d *DescriptorProto, node *generatorNode) {
	var chs []x

	for i, m := range d.NestedType {
		chs = append(chs, x{
			Data: m,
			Path: fmt.Sprintf("%s,%d,%d", path, 3, i),
			Type: sortTypeMessage,
			Name: m.GetName(),
		})
	}

	for i, e := range d.EnumType {
		chs = append(chs, x{
			Data: e,
			Path: fmt.Sprintf("%s,%d,%d", path, 4, i),
			Type: sortTypeEnum,
			Name: e.GetName(),
		})
	}

	for i, r := range d.ReservedRange {
		chs = append(chs, x{
			Data:  r,
			Path:  fmt.Sprintf("%s,%d,%d", path, 9, i),
			Type:  sortTypeNumber,
			Field: r.Start,
		})
	}

	for i, r := range d.ReservedName {
		chs = append(chs, x{
			Data: r,
			Path: fmt.Sprintf("%s,%d,%d", path, 10, i),
			Type: sortTypeReserved,
			Name: r,
		})
	}

	for i, f := range d.Field {
		if f.OneofIndex == nil {
			chs = append(chs, x{
				Data:  f,
				Path:  fmt.Sprintf("%s,%d,%d", path, 2, i),
				Type:  sortTypeNumber,
				Field: f.Number,
			})
		} else {
			chs = append(chs, x{
				Data: oow{
					Field: f,
					Oneof: d.OneofDecl[*f.OneofIndex],
				},
				Path:  fmt.Sprintf("%s,%d,%d", path, 2, i),
				Type:  sortTypeNumber,
				Field: f.Number,
			})
		}
	}

	sort.Slice(chs, func(i, j int) bool {
		ci := chs[i]
		cj := chs[j]
		if ci.Type == cj.Type {
			if ci.Type == sortTypeNumber {
				return *ci.Field < *cj.Field
			}
			return ci.Name < cj.Name
		}
		return ci.Type < cj.Type
	})
	//for i, c := range chs {
	//	v, ok := c.Data.(string)
	//	if !ok {
	//		v = c.Data.(fmt.Stringer).String()
	//	}
	//	ot.E("li").T(strconv.Itoa(i+1), ". ", v)
	//}

	g.generateComment(path, node)
	node.E("span", html.Attribute{Key: "class", Val: "keyword"}, html.Attribute{Key: "id", Val: d.GetName()}).T("message")
	node.T(" ", d.GetName(), " {")
	ot := node.E("ol")
	s := make(map[*FieldDescriptorProto]struct{})
	for _, c := range chs {
		switch v := c.Data.(type) {
		case *FieldDescriptorProto:
			g.generateField(c.Path, v, ot.E("li"))
		case oow:
			if _, ok := s[v.Field]; ok {
				continue
			}
			lt := ot.E("li")
			lt.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("oneof")
			lt.T(" ", v.Oneof.GetName(), " {")
			for _, oc := range chs {
				f, ok := oc.Data.(oow)
				if !ok {
					continue
				}
				s[f.Field] = struct{}{}
				g.generateField(c.Path, f.Field, lt.E("ol").E("li"))
			}
			lt.T("}")
		case *EnumDescriptorProto:
			g.generateEnum(indent, c.Path, v, ot.E("li", html.Attribute{Key: "class", Val: "struct"}))
		case *DescriptorProto_ReservedRange:
			g.generateReservedRange(c.Path, v, ot.E("li"))
		case *DescriptorProto:
			g.generateMessage(f, indent, c.Path, v, ot.E("li", html.Attribute{Key: "class", Val: "struct"}))
		case string:
			g.generateReservedName(c.Path, v, ot.E("li"))
		default:
			panic(fmt.Sprintf("invalid type: %T", v))
		}
	}
	node.T("}")
}

func (g *generator) generateComment(path string, node *generatorNode) {
	loc, ok := g.comment(path)
	if !ok {
		return
	}
	s := node.E("span", html.Attribute{Key: "class", Val: "comment"})
	for _, l := range strings.Split(strings.Trim(*loc.LeadingComments, "\n"), "\n") {
		s.T("//", l)
		s.E("br")
	}
}

func (g *generator) generateReservedRange(path string, r *DescriptorProto_ReservedRange, node *generatorNode) {
	g.generateComment(path, node)
	node.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("reserved")
	node.T(" ")
	node.E("span", html.Attribute{Key: "class", Val: "value"}).T(r.GetStart())
	if r.GetStart()+1 < r.GetEnd() {
		node.T(r.GetEnd())
		node.E("span", html.Attribute{Key: "class", Val: "value"}).T(r.GetEnd())
	}
	node.T(";")
}

func (g *generator) generateReservedName(path string, name string, node *generatorNode) {
	g.generateComment(path, node)
	node.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("reserved")
	node.T(" ")
	node.E("span", html.Attribute{Key: "class", Val: "value"}).T(`"`, name, `"`)
	node.T(";")
}

func (g *generator) generateField(path string, field *FieldDescriptorProto, node *generatorNode) {
	g.generateComment(path, node)
	g.addTypeLink(node, field)
	node.T(" ", field.GetName(), " = ")
	node.E("span", html.Attribute{Key: "class", Val: "value"}).T(field.GetNumber())
	node.T(";")
}

func (g *generator) generateEnum(indent, path string, enum *EnumDescriptorProto, node *generatorNode) {
	g.generateComment(path, node)
	node.E("span", html.Attribute{Key: "class", Val: "keyword struct"}, html.Attribute{Key: "id", Val: enum.GetName()}).T("enum")
	node.T(" ", enum.GetName(), " {")
	ut := node.E("ul").E("li").E("table")
	for i, v := range enum.GetValue() {
		g.generateComment(fmt.Sprintf("%s,%d,%d", path, 2, i), ut.E("tr").E("td", html.Attribute{Key: "colspan", Val: "3"}))
		node := ut.E("tr")
		node.E("td").T(v.GetName())
		node.E("td").T(" = ")
		vt := node.E("td", html.Attribute{Key: "text-align", Val: "right"})
		vt.E("span", html.Attribute{Key: "class", Val: "value"}).T(v.GetNumber())
		vt.T(";")
		// todo: options
	}
	node.T("}")
}

func (g *generator) pommes(indent string, node *generatorNode, path string, d *DescriptorProto, f *FileDescriptorProto) {
	indent = "\t"

	pt := node.E("li", html.Attribute{Key: "id", Val: d.GetName()}).E("code").E("pre")

	loc, ok := g.comment(path)
	if ok {
		pt.E("span", html.Attribute{Key: "class", Val: "comment"}).T("//", strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n//", -1), "\n")
	}
	pt.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("message")
	pt.T(" ", d.GetName(), " {\n")

	if len(d.NestedType) > 0 {
		ut := pt.E("ul")
		for i, message := range d.NestedType {
			g.generateMessage(f, indent+"\t", fmt.Sprintf("%s,3,%d", path, i), message, ut)
		}
	}

	if len(d.EnumType) > 0 {
		for _, x := range d.EnumType {
			pt.T("enum", " ", x.GetName(), " {\n")
			for _, y := range x.Value {
				pt.T(indent, y.GetName(), " = ", y.GetNumber(), "\n")
			}
			pt.T("}\n")
		}
	}

	if len(d.OneofDecl) > 0 {
		for _, x := range d.OneofDecl {
			pt.T("oneof", " ", x.GetName(), "\n")
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
		pt.T(indent)
		pt.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("reserved")
		pt.T(" ", strings.Join(rs, ", "), ";\n")
	}

	if len(d.GetReservedName()) > 0 {
		var rs []string
		for _, x := range d.GetReservedName() {
			rs = append(rs, `"`+x+`"`)
		}
		pt.T(indent)
		pt.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("reserved")
		pt.T(" ", strings.Join(rs, ", "), ";\n")
	}

	for i, field := range d.Field {
		loc, ok := g.comment(fmt.Sprintf("%s,2,%d", path, i))
		if ok {
			if i != 0 {
				pt.T("\n")
			}
			pt.T(indent)
			pt.E("span", html.Attribute{Key: "class", Val: "comment"}).T("//", strings.Replace(strings.Trim(*loc.LeadingComments, "\n"), "\n", "\n"+indent+"//", -1), "\n")
		}

		pt.T(indent)

		g.addTypeLink(pt, field)

		//pt.T(fmt.Sprintf("%d %#v", field.GetOneofIndex(), field))

		pt.T(" ", field.GetName(), " = ")
		pt.E("span", html.Attribute{Key: "class", Val: "value"}).T(strconv.Itoa(int(field.GetNumber())))

		if field.Options != nil {
			pt.T(" [\n")
			if field.Options.GetDeprecated() {
				pt.T(indent, "\t")
				pt.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("deprecated")
				pt.T(" = ")
				pt.E("span", html.Attribute{Key: "class", Val: "value"}).T(true)
				pt.T("\n")
			}
			pt.T(indent, "];")
		}

		pt.T("\n")
	}
	pt.T(indent[:len(indent)-1], "}")

	//u := pt.E("ul")
	//for _, c := range chs {
	//	u.E("li").T(fmt.Sprintf("%T", c))
	//}
}

func (g *generator) addTypeLink(node *generatorNode, field *FieldDescriptorProto) {
	tnp := strings.Split(field.GetTypeName(), ".")[1:]
	switch len(tnp) {
	case 0:
		node.E("span", html.Attribute{Key: "class", Val: "type"}).T(typeMap[field.GetType()])
	case 1:
		node.E("span", html.Attribute{Key: "class", Val: "type"}).T(field.GetTypeName())
	default:
		o := len(tnp) - 2
		if o < 0 {
			o = 0
		}
		tl := strings.Replace(g.typeToFile[field.GetTypeName()[1:]], ".proto", ".html", 1) + "#" + tnp[len(tnp)-1]
		node.E("a", html.Attribute{Key: "href", Val: g.linkBase + tl}).E("span", html.Attribute{Key: "class", Val: "type"}).T(strings.Join(tnp[o:], "."))
	}
}

func normalizeTypeName(field *FieldDescriptorProto) string {
	tnp := strings.Split(field.GetTypeName(), ".")[1:]
	switch len(tnp) {
	case 0:
		return typeMap[field.GetType()]
	case 1:
		return field.GetTypeName()
	default:
		o := len(tnp) - 2
		if o < 0 {
			o = 0
		}
		return strings.Join(tnp[o:], ".")
	}
}
