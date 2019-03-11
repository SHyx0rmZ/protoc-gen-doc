package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
)

type generator struct {
	*bytes.Buffer
	typeToFile map[string]string
	linkBase   string
	comments   map[string][]*SourceCodeInfo_Location
}

type generatorNode struct {
	*html.Node
}

func (g *generatorNode) E(data string, attrs ...html.Attribute) *generatorNode {
	node := &html.Node{
		Type: html.ElementNode,
		Data: data,
		Attr: attrs,
	}
	g.AppendChild(node)
	return &generatorNode{node}
}

func (g *generatorNode) T(data ...interface{}) *generatorNode {
	node := &html.Node{
		Type: html.TextNode,
		Data: fmt.Sprint(data...),
	}
	g.AppendChild(node)
	return &generatorNode{node}
}
