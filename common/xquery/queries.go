package xquery

import (
	"bytes"

	xq "github.com/antchfx/xmlquery"
)

// XML is a wrapper of xmlquery.XML. It has additional safety incorporated when querying the contents.
// This also serves as an abstraction from real implementation which can change.
// In fact library was changed, this should limit disruption in the future if such issue arises again.
type XML struct {
	delegate *xq.Node
}

func NewXML(data []byte) (*XML, error) {
	node, err := xq.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return newXML(node), nil
}

func newXML(node *xq.Node) *XML {
	return &XML{delegate: node}
}

func (x *XML) IsEmpty() bool {
	// this check allows to abide to Null Object pattern
	return x.delegate == nil
}

func (x *XML) FindOne(expr string) (output *XML) {
	if x.IsEmpty() {
		return x
	}

	defer panicRecovery(func(cause error) { // nolint:unparam
		output = newXML(nil)
	})

	node := xq.FindOne(x.delegate, expr)

	return newXML(node)
}

func (x *XML) FindMany(expr string) (output []*XML) {
	if x.IsEmpty() {
		return []*XML{}
	}

	defer panicRecovery(func(cause error) { // nolint:unparam
		output = []*XML{}
	})

	nodes := xq.Find(x.delegate, expr)
	output = make([]*XML, len(nodes))

	for i, node := range nodes {
		output[i] = newXML(node)
	}

	return output
}

func (x *XML) Parent() *XML {
	if x.IsEmpty() {
		return x
	}

	return newXML(x.delegate.Parent)
}

func (x *XML) Attr(name string) string {
	if x.IsEmpty() {
		return ""
	}

	return x.delegate.SelectAttr(name)
}

func (x *XML) HasChildren() bool {
	return x.delegate.FirstChild != nil
}

func panicRecovery(wrapup func(cause error)) {
	if re := recover(); re != nil {
		err, ok := re.(error)
		if !ok {
			panic(re)
		}

		wrapup(err)
	}
}
