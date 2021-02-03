package data_struct

import (
	"fmt"
	"strings"
)

//树状结构存储

type Node struct {
	tag    string
	child  map[string]*Node
	attrs  map[string]string
	parent *Node
}

func NewNode(t string) *Node {
	n := new(Node)
	n.tag = t
	n.attrs = make(map[string]string)
	n.child = make(map[string]*Node)
	return n
}

func dump(str *[]string, node *Node) {
	*str = append(*str, node.String())
	for _, v := range node.child {
		dump(str, v)
	}
}

func (n *Node) Root() *Node {
	var p, it *Node
	for p = n.parent; p != nil; p = p.parent {
		if p.tag != "" {
			it = p
		}
	}

	return it
}

func (n *Node) Update(kv map[string]string) {
	for k, v := range kv {
		n.attrs[k] = v
	}
}

func (n *Node) Remove(it string) {
	delete(n.child, it)
}

func (n *Node) AddChild(node *Node) {
	n.child[node.tag] = node
}

func (n *Node) String() string {
	var ret []string
	for k, v := range n.attrs {
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("%s;%s", n.tag, strings.Join(ret, ";"))
}
