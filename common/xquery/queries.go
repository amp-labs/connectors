package xquery

import (
	"bytes"
	"errors"

	xq "github.com/antchfx/xmlquery"
)

var ErrInvalidXML = errors.New("data is not in xml format")

// XML is a wrapper of xmlquery.XML. It has additional safety incorporated when querying the contents.
// This also serves as an abstraction from real implementation which can change.
// In fact library was changed, this should limit disruption in the future if such issue arises again.
type XML struct {
	delegate *xq.Node
}

func NewXML(data []byte) (*XML, error) {
	node, err := xq.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Join(err, ErrInvalidXML)
	}

	// try to create XML node again, but now trim any spaces
	withoutEmptySpaces := []byte(node.OutputXML(true))

	node, err = xq.Parse(bytes.NewReader(withoutEmptySpaces))
	if err != nil {
		return nil, errors.Join(err, ErrInvalidXML)
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
		return nil
	}

	defer panicRecovery(func(cause error) { // nolint:unparam
		output = nil
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

// RawXML returns raw data of the whole XML node.
func (x *XML) RawXML() string {
	if x.IsEmpty() {
		return ""
	}

	return x.delegate.OutputXML(true)
}

// Text provides inner text of this node.
func (x *XML) Text() string {
	if x.IsEmpty() {
		return ""
	}

	return x.delegate.InnerText()
}

// SetDataText verifies that data type matches the type of xml Node.
// Changes text.
func (x *XML) SetDataText(data string) {
	if x.IsEmpty() {
		return
	}

	if x.delegate.Type == xq.TextNode {
		x.delegate.Data = data
	}
}

// SetDataNode will attach tree found under the node to current xml as a First element.
// If this xml object cannot accept then nodes nothing happens.
// Returns true if set operation was successful.
func (x *XML) SetDataNode(node *XML) bool {
	// attach XML tree as the first child
	if x.IsEmpty() || node.IsEmpty() {
		return false
	}

	// current xml can accept element node if it is of such type.
	if x.delegate.Type == xq.ElementNode {
		// Tree will have wrapping tags which are not relevant.
		// The very first element node is a root of linked list.
		// Set this root to adopt all children in the list.
		for _, child := range node.GetChildren() {
			if child.delegate.Type == xq.ElementNode {
				x.delegate.FirstChild = child.delegate

				return true
			}
		}
	}

	return false
}

func (x *XML) GetChild() *XML {
	if x.IsEmpty() {
		return newXML(nil)
	}

	return newXML(x.delegate.FirstChild)
}

func (x *XML) GetChildren() []*XML {
	firstChild := x.GetChild()
	if firstChild.IsEmpty() {
		return nil
	}

	result := make([]*XML, 0)

	// follow next child to collect them in list
	next := firstChild.delegate
	for next != nil {
		result = append(result, newXML(next))
		next = next.NextSibling
	}

	return result
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
