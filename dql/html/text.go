/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package html

import (
	"github.com/goplus/xgo/dql"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// -----------------------------------------------------------------------------

// NodeFilter is the interface for filtering nodes when retrieving text content.
type NodeFilter interface {
	// Filter returns true if the node should be included in the text content.
	Filter(*html.Node) bool

	// TextNodeData returns the text data of a text node. It can be used to customize
	// the text data of a text node, for example, to trim spaces or to replace certain
	// characters.
	TextNodeData(*html.Node) string
}

type noFilter struct{}

func (f noFilter) Filter(*html.Node) bool { return true }
func (f noFilter) TextNodeData(node *html.Node) string {
	return node.Data
}

// textOf returns text data of all node's children.
func textOf[F NodeFilter](node *html.Node, outer bool, f F) string {
	p := textPrinter[F]{
		nodef: f,
	}
	p.printNode(node, outer, false)
	return string(p.data)
}

type textPrinter[F NodeFilter] struct {
	data         []byte
	nodef        F
	notLineStart bool
	hasSpace     bool
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func (p *textPrinter[F]) printCollapsed(v string) {
	for len(v) > 0 {
		n := len(v)
		i := 0
		for i < n && isSpace(v[i]) {
			i++ // skip leading spaces
		}
		if i > 0 {
			p.hasSpace = true
		}
		if i >= n {
			break
		}
		if p.notLineStart && p.hasSpace {
			p.data = append(p.data, ' ')
		} else {
			p.notLineStart = true
		}
		p.hasSpace = false
		start := i
		i++
		for i < n && !isSpace(v[i]) {
			i++
		}
		p.data = append(p.data, v[start:i]...)
		v = v[i:]
	}
}

func (p *textPrinter[F]) printVerbatim(v string) {
	p.data = append(p.data, v...)
	if len(v) > 0 {
		last := v[len(v)-1]
		p.notLineStart = last != '\n'
		p.hasSpace = p.notLineStart && isSpace(last)
	}
}

func (p *textPrinter[F]) printNode(node *html.Node, outer, verbatim bool) {
	if node == nil {
		return
	}
	f := p.nodef
	if node.Type == html.TextNode {
		data := f.TextNodeData(node)
		if verbatim {
			p.printVerbatim(data)
		} else {
			p.printCollapsed(data)
		}
		return
	}
	verbatim = verbatim || node.DataAtom == atom.Pre
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if f.Filter(child) {
			p.printNode(child, true, verbatim)
		}
	}
	if outer {
		switch node.DataAtom {
		case atom.P, atom.Div, atom.Br, atom.H1, atom.H2, atom.H3, atom.H4,
			atom.H5, atom.H6, atom.Li, atom.Blockquote, atom.Pre:
			p.data = append(p.data, '\n')
			p.notLineStart = false
		}
	}
}

// -----------------------------------------------------------------------------

// Text retrieves the text content of the NodeSet. It only retrieves from the
// first node in the NodeSet. It ignores any error and returns an empty string
// if there is an error.
func (p NodeSet) Text__0() string {
	val, _ := p.Text__1()
	return val
}

// Text retrieves the text content of the NodeSet. It only retrieves from the
// first node in the NodeSet.
func (p NodeSet) Text__1() (val string, err error) {
	node, err := p.First()
	if err == nil {
		val = textOf(&node.Node, false, noFilter{})
	}
	return
}

// Text retrieves the text content of the NodeSet with a node filter. It only
// retrieves from the first node in the NodeSet.
//
// The node filter is used to filter the nodes when retrieving text content. If
// the node filter returns false for a node, the node and its children will be
// ignored when retrieving text content.
//
// The outer parameter specifies whether to include the text content of the
// outer node itself. If outer is true, the text content of the outer node will
// be included; otherwise, only the text content of the inner nodes will be
// included.
func Text[F NodeFilter](ns NodeSet, outer bool, f F) (val string, err error) {
	node, err := ns.First()
	if err == nil {
		val = textOf(&node.Node, outer, f)
	}
	return
}

// Int retrieves the integer value from the text content of the first node in
// the NodeSet.
func (p NodeSet) Int() (int, error) {
	text, err := p.Text__1()
	if err != nil {
		return 0, err
	}
	return dql.Int(text)
}

// -----------------------------------------------------------------------------
