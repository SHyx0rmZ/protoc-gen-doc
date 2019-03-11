package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/html"
	"sort"
	"strings"
)

func (g *generator) generateService(s *ServiceDescriptorProto, path string, node *generatorNode) {
	pt := node.E("li").E("code").E("pre")
	pt.T("service ", s.GetName(), " {\n")

	sort.Slice(s.GetMethod(), func(i, j int) bool {
		return s.GetMethod()[i].GetName() < s.GetMethod()[j].GetName()
	})
	for i, m := range s.GetMethod() {
		if i != 0 {
			pt.T("\n")
		}
		loc, ok := g.comment(fmt.Sprintf("%s,2,%d", path, i))
		if ok {
			pt.E("span", html.Attribute{Key: "class", Val: "comment"}).T("\t//", strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n\t//", -1), "\n")
		}
		pt.T("\trpc ", m.GetName(), "(")
		if m.GetClientStreaming() {
			pt.T("stream ")
		}
		g.addTypeLink(pt, &FieldDescriptorProto{TypeName: proto.String(m.GetInputType())})
		pt.T(") returns (")
		if m.GetServerStreaming() {
			pt.T("stream ")
		}
		g.addTypeLink(pt, &FieldDescriptorProto{TypeName: proto.String(m.GetOutputType())})
		pt.T(")")
		if m.Options != nil {
			pt.T(" {\n")
			if m.Options.GetDeprecated() {
				pt.T("\t\t", "option deprecated = true", ";\n")
			}
			pt.T("\t}")
		}
		pt.T(";\n")
	}
	pt.T("}")
}
