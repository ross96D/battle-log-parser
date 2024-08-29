package parser

import "golang.org/x/net/html"

func attr(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func find(n *html.Node, condition func(n *html.Node) bool) *html.Node {
	if condition(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res := find(c, condition)
		if res != nil {
			return res
		}
	}
	return nil
}

func findAll(n *html.Node, condition func(n *html.Node) bool, result []*html.Node) []*html.Node {
	if result == nil {
		result = make([]*html.Node, 0)
	}

	if condition(n) {
		result = append(result, n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = findAll(c, condition, result)
	}

	return result
}

func forEach(n *html.Node, f func(n *html.Node)) {
	f(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEach(c, f)
	}
}
