package readability

// html.Node 的简单包装
import (
	"bytes"
	"io"

	"golang.org/x/net/html"
)

// Node Wrap of Node
type Node struct {
	*html.Node
	// Attr map[string]string
}

func (node *Node) ParentNode() *Node {
	if node != nil {
		return &Node{node.Parent}
	}
	return nil
}

func NewDocument(r io.Reader) (root *Node, err error) {
	n, err := html.Parse(r)
	if err != nil {
		return
	}
	root = &Node{n}
	return
}

func (node *Node) IsComment() bool {
	if node.Type == html.CommentNode {
		return true
	}
	return false
}

func (node *Node) GetTagName() string {
	if node.Type == html.ElementNode {
		return node.Data
	}
	return ""
}

func (node *Node) GetAttr(attr string) string {
	for _, a := range node.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}
func (node *Node) GetAttrOr(attr, or string) string {
	for _, a := range node.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return or
}

// Work work the dom node
func (node *Node) Work(f func(child *Node)) {
	if node == nil || node.Node == nil {
		return
	}
	f(node)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		cNode := &Node{c}
		cNode.Work(f)
	}
}

// WorkTextNode work textnode
func (node *Node) WorkTextNode(f func(child *Node)) {
	if node == nil || node.Node == nil {
		return
	}

	f(node)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			cNode := &Node{c}
			cNode.WorkTextNode(f)
		}
	}
}

// WorkElementNode work the dom node by tag
func (node *Node) WorkElementNode(tag string, f func(child *Node)) {
	if node == nil || node.Node == nil {
		return
	}

	f(node)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			cNode := &Node{c}
			cNode.WorkElementNode(tag, f)
		}
	}
}

// InsertTextBefore Insert text before
func (node *Node) InsertTextBefore(text string) (textNode *Node) {
	textNode = NewNode(html.TextNode)
	textNode.Data = text
	node.InsertBefore(textNode)
	return
}

// InsertTextAfter insert Text after
func (node *Node) InsertTextAfter(text string) (textNode *Node) {
	textNode = NewNode(html.TextNode)
	textNode.Data = text
	node.InsertBefore(textNode)

	return
}

// InsertElementBefore insert element node before
func (node *Node) InsertElementBefore(tag string) (elementNode *Node) {
	elementNode = NewNode(html.ElementNode)
	elementNode.Data = tag

	node.InsertBefore(elementNode)
	return
}

// InsertElementAfter insert element node after
func (node *Node) InsertElementAfter(tag string) (elementNode *Node) {
	elementNode = NewNode(html.ElementNode)
	elementNode.Data = tag

	node.InsertAfter(elementNode)
	return

}

func (node *Node) InsertAfter(newNode *Node) {
	newNode.Parent = node.Parent
	newNode.NextSibling = node.Node
	if node.PrevSibling != nil {
		node.PrevSibling.NextSibling = newNode.Node
		node.PrevSibling = newNode.Node
	} else if node.Parent != nil {
		node.Parent.FirstChild = newNode.Node
	}

	return
}

func (node *Node) InsertBefore(newNode *Node) {
	newNode.Parent = node.Parent
	newNode.NextSibling = node.Node
	if node.PrevSibling != nil {
		node.PrevSibling.NextSibling = newNode.Node
		node.PrevSibling = newNode.Node
	} else if node.Parent != nil {
		node.Parent.FirstChild = newNode.Node
	}

	return
}

func (node *Node) InnerText() string {
	var buf bytes.Buffer

	node.WorkTextNode(func(n *Node) {
		buf.WriteString(n.Data)
	})

	return buf.String()
}

func (node *Node) Html() (res string, e error) {
	var buf bytes.Buffer

	e = html.Render(&buf, node.Node)
	if e != nil {
		return
	}
	res = buf.String()
	return
}

func (node *Node) Remove() {
	if node.Parent != nil {
		node.Parent.RemoveChild(node.Node)

		n := node.Parent
		c := node.Node

		if n != nil && n.FirstChild == c {
			n.FirstChild = c.NextSibling
		}
		if c.NextSibling != nil {
			c.NextSibling.PrevSibling = c.PrevSibling
		}
		if n != nil && n.LastChild == c {
			n.LastChild = c.PrevSibling
		}
		if c.PrevSibling != nil {
			c.PrevSibling.NextSibling = c.NextSibling
		}

		c.Parent = nil
		c.PrevSibling = nil
	}
}

func (node *Node) AppendChild(child *Node) {
	node.Node.AppendChild(child.Node)
}

func (node *Node) TagLength(tag string) int {
	l := 0
	node.Work(func(n *Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			l++
		}
	})
	return l
}

func NewNode(t html.NodeType) *Node {
	m := &Node{
		Node: &html.Node{
			Type: t,
			Attr: make([]html.Attribute, 0),
		},
	}
	return m
}
