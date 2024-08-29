package parser

import (
	"slices"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func removeEmptyLines(lines []string) []string {
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func parseNumBackwards(s string, start int) (uint64, error) {
	numStr := []byte{}
	for i := start; i >= 0; i-- {
		char := s[i]
		if char == ' ' {
			break
		}
		numStr = append(numStr, char)
	}
	slices.Reverse(numStr)
	return strconv.ParseUint(string(numStr), 10, 64)
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
