package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"strings"
)

func (g *generator) generateService(s *ServiceDescriptorProto, path string, node *generatorNode) {
	pt := node.E("li").E("code").E("pre")
	pt.T("service ", s.GetName(), " {\n")

	for i, m := range s.GetMethod() {
		if i != 0 {
			pt.T("\n")
		}
		loc, ok := g.comment(fmt.Sprintf("%s,2,%d", path, i))
		if ok {
			pt.T("\t//", strings.Replace(strings.TrimSuffix(*loc.LeadingComments, "\n"), "\n", "\n\t//", -1), "\n")
		}
		pt.T("\trpc ", m.GetName(), "(")
		if m.GetClientStreaming() {
			pt.T("stream ")
		}
		pt.T(normalizeTypeName(&FieldDescriptorProto{TypeName: proto.String(m.GetInputType())}), ") returns (")
		if m.GetServerStreaming() {
			pt.T("stream ")
		}
		pt.T(normalizeTypeName(&FieldDescriptorProto{TypeName: proto.String(m.GetOutputType())}), ")")
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
