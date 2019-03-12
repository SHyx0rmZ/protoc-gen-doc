package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/html"
	"sort"
)

func (g *generator) generateService(s *ServiceDescriptorProto, path string, node *generatorNode) {
	g.generateLeadingComments(path, node)
	node.E("span", html.Attribute{Key: "class", Val: "keyword"}, html.Attribute{Key: "id", Val: s.GetName()}).T("service")
	node.T(" ", s.GetName(), " {")
	ot := node.E("ol")

	var ms []struct {
		Method *MethodDescriptorProto
		Index  int
	}
	for i, m := range s.GetMethod() {
		ms = append(ms, struct {
			Method *MethodDescriptorProto
			Index  int
		}{
			Method: m,
			Index:  i,
		})
	}
	sort.Slice(ms, func(i, j int) bool {
		return ms[i].Method.GetName() < ms[j].Method.GetName()
	})
	for _, m := range ms {
		g.generateMethod(fmt.Sprintf("%s,%d,%d", path, 2, m.Index), m.Method, ot.E("li"))
	}
	node.T("}")
	g.generateTrailingComments(path, node)
}

func (g *generator) generateMethod(path string, m *MethodDescriptorProto, node *generatorNode) {
	g.generateLeadingComments(path, node)
	node.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("rpc")
	node.T(" ", m.GetName(), "(")
	if m.GetClientStreaming() {
		node.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("stream")
		node.T(" ")
	}
	g.addTypeLink(node, &FieldDescriptorProto{TypeName: proto.String(m.GetInputType())})
	node.T(") returns (")
	if m.GetServerStreaming() {
		node.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("stream")
		node.T(" ")
	}
	g.addTypeLink(node, &FieldDescriptorProto{TypeName: proto.String(m.GetOutputType())})
	node.T(")")
	if m.Options != nil {
		node.T(" {")
		ut := node.E("ul")
		if m.Options.GetDeprecated() {
			lt := ut.E("li")
			lt.E("span", html.Attribute{Key: "class", Val: "keyword"}).T("option")
			lt.T(" ", "deprecated", " = ")
			lt.E("span", html.Attribute{Key: "class", Val: "value"}).T("true")
			lt.T(";")
		}
		node.T("}")
	}
	node.T(";")
	g.generateTrailingComments(path, node)
}
