package xquery

import (
	"strings"

	xq "github.com/antchfx/xmlquery"
)

func (x *XML) EqualsIgnoreOrder(other *XML) bool {
	return xmlNodesEqual(x.delegate, other.delegate)
}

func xmlNodesEqual(first, second *xq.Node) bool { // nolint:cyclop,funlen
	if first == nil && second == nil {
		return true
	}

	if first == nil || second == nil {
		return false
	}

	// Compare element names
	if first.Data != second.Data {
		return false
	}

	// Compare namespace
	if first.Prefix != second.Prefix || first.NamespaceURI != second.NamespaceURI {
		return false
	}

	// Compare attributes ignoring order
	if !equalXMLAttrs(first.Attr, second.Attr) {
		return false
	}

	// Gather element children
	firstElemChildren := childElements(first)
	secondElemChildren := childElements(second)

	// If there are no element children on either side, compare leaf text (trimmed)
	if len(firstElemChildren) == 0 && len(secondElemChildren) == 0 {
		text1 := strings.TrimSpace(first.InnerText())
		text2 := strings.TrimSpace(second.InnerText())

		return text1 == text2
	}

	// If either side has significant text nodes (non-whitespace) *and* element children,
	// we must compare the sequence of significant child nodes in order.
	if hasNonWhitespaceTextChild(first) || hasNonWhitespaceTextChild(second) {
		return orderedSignificantChildrenEqual(first, second)
	}

	// At this point: both nodes have element children and no significant text nodes.
	// Compare element children as unordered multisets (existing logic).

	if len(firstElemChildren) != len(secondElemChildren) {
		return false
	}

	used := make([]bool, len(secondElemChildren))

	for _, fc := range firstElemChildren {
		matchFound := false

		for i, sc := range secondElemChildren {
			if used[i] {
				continue
			}

			if xmlNodesEqual(fc, sc) {
				used[i] = true
				matchFound = true

				break
			}
		}

		if !matchFound {
			return false
		}
	}

	return true
}

// hasNonWhitespaceTextChild returns true if node has any text child with non-whitespace content.
func hasNonWhitespaceTextChild(n *xq.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xq.TextNode && strings.TrimSpace(child.Data) != "" {
			return true
		}
	}

	return false
}

// orderedSignificantChildrenEqual compares the sequence of significant children (text nodes
// with non-whitespace and element nodes) in order. Text children are compared trimmed.
func orderedSignificantChildrenEqual(firstNode, secondNode *xq.Node) bool {
	// Build slices of significant nodes for both
	children1 := significantChildren(firstNode)
	children2 := significantChildren(secondNode)

	if len(children1) != len(children2) {
		return false
	}

	for i := range children1 {
		child1 := children1[i]
		child2 := children2[i]

		// Both text nodes -> compare trimmed content
		if child1.Type == xq.TextNode && child2.Type == xq.TextNode {
			if strings.TrimSpace(child1.Data) != strings.TrimSpace(child2.Data) {
				return false
			}

			continue
		}

		// Both element nodes -> recurse
		if child1.Type == xq.ElementNode && child2.Type == xq.ElementNode {
			if !xmlNodesEqual(child1, child2) {
				return false
			}

			continue
		}

		// Different types (one text, one element) -> not equal
		return false
	}

	return true
}

// significantChildren returns child nodes that are either element nodes or non-whitespace text nodes,
// preserving document order.
func significantChildren(n *xq.Node) []*xq.Node {
	var result []*xq.Node

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xq.ElementNode ||
			(child.Type == xq.TextNode && strings.TrimSpace(child.Data) != "") {
			result = append(result, child)
		}
	}

	return result
}

func childElements(n *xq.Node) []*xq.Node {
	var children []*xq.Node

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xq.ElementNode {
			children = append(children, child)
		}
	}

	return children
}

func equalXMLAttrs(first, second []xq.Attr) bool {
	if len(first) != len(second) {
		return false
	}

	// Create lookup from the first attribute list
	lookup := make(map[string]string, len(first))

	for _, at := range first {
		k := at.Name.Space + ":" + at.Name.Local
		lookup[k] = at.Value
	}

	// Compare against the second
	for _, attr := range second {
		key := attr.Name.Space + ":" + attr.Name.Local
		if v, ok := lookup[key]; !ok || v != attr.Value {
			return false
		}
	}

	return true
}
