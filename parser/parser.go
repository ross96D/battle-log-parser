package parser

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func Parse(data io.ReadCloser) (b Battle, err error) {
	root, err := html.Parse(data)
	if err != nil {
		return
	}
	data.Close()

	body := find(root, func(n *html.Node) bool {
		return n.Data == "body"
	})

	cardList := findAll(body, func(n *html.Node) bool {
		if value, ok := attr(n, "class"); ok {
			return value == "card"
		}
		return false
	}, nil)

	resumeNode := cardList[0]
	identifierNode := cardList[1]
	_ = identifierNode

	b.Date, err = ParseIdentifierNode(cardList[1])
	if err != nil {
		return Battle{}, err
	}

	b.Resume = ParseResumeNode(resumeNode)
	b.Turns = ParseTurnNodes(cardList[2 : len(cardList)-1])

	return
}

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

func getNodeLines(n *html.Node) []string {
	b := strings.Builder{}
	forEach(n, func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		if n.Type == html.ElementNode {
			if n.Data == "br" {
				b.WriteString("\n")
			}
		}
	})

	lines := strings.Split(b.String(), "\n")
	lines = removeEmptyLines(lines)

	return lines
}
