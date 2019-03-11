package main

import (
	"fmt"
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

func (g *generator) generateMessage(f *FileDescriptorProto, indent, path string, d *DescriptorProto, node *generatorNode) {
	indent = "\t"

	pt := node.E("li", html.Attribute{Key: "id", Val: d.GetName()}).E("code").E("pre")

	loc, ok := g.comment(path)
	if ok {
		pt.T("//", strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n//", -1), "\n")
	}
	pt.T("message ", d.GetName(), " {\n")

	if len(d.NestedType) > 0 {
		ut := pt.E("ul")
		for i, message := range d.NestedType {
			g.generateMessage(f, indent+"\t", fmt.Sprintf("%s,3,%d", path, i), message, ut)
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
		pt.T(indent, "reserved ", strings.Join(rs, ", "), ";\n")
	}

	if len(d.GetReservedName()) > 0 {
		var rs []string
		for _, x := range d.GetReservedName() {
			rs = append(rs, `"`+x+`"`)
		}
		pt.T(indent, "reserved ", strings.Join(rs, ", "), ";\n")
	}

	for i, field := range d.Field {

		pt.T(indent)

		loc, ok := g.comment(fmt.Sprintf("%s,2,%d", path, i))
		if ok {
			pt.T("//", strings.Replace(strings.Trim(*loc.LeadingComments, "\n"), "\n", "\n"+indent+"//", -1), "\n", indent)
		}

		pt.T(normalizeTypeName(field))

		//if len(field.GetTypeName()) == 0 {
		//	g.P(fmt.Sprintf("%T %#v", field.GetType(), field.GetType()))
		//}
		//
		pt.T(" ", field.GetName(), " = ", strconv.Itoa(int(field.GetNumber())))

		if field.Options != nil {
			pt.T(" [\n")
			if field.Options.GetDeprecated() {
				pt.T(indent, "\t", "deprecated = true", "\n")
			}
			pt.T(indent, "];")
		}

		pt.T("\n\n")
	}
	pt.T(indent[:len(indent)-1], "}")
}

func normalizeTypeName(field *FieldDescriptorProto) string {
	tn := field.GetTypeName()
	tnp := strings.Split(tn, ".")[1:]
	switch len(tnp) {
	case 0:
		return typeMap[field.GetType()]
	case 1:
		return tn
	default:
		o := len(tnp) - 2
		if o < 0 {
			o = 0
		}
		return strings.Join(tnp[o:], ".")
	}
}
