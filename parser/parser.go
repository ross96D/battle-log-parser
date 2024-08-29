package parser

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"
	"time"

	"github.com/ross96D/battle-log-parser/assert"
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

func ParseIdentifierNode(n *html.Node) (time.Time, error) {
	lines := getNodeLines(n)
	assert.Assert(len(lines) == 1, "IdentifierNode have only one line %d", len(lines))
	line := lines[0]
	hour := []byte{}
	date := []byte{}
	onHour := true
	for i := len(line) - 1; i >= 0; i-- {
		char := line[i]
		if onHour {
			if char == ' ' {
				onHour = false
				continue
			}
			hour = append(hour, char)
		} else {
			if char == ' ' {
				break
			}
			date = append(date, char)
		}
	}
	slices.Reverse(date)
	slices.Reverse(hour)
	hourSplitted := bytes.Split(hour, []byte(":"))
	dateSplitted := bytes.Split(date, []byte("-"))

	hourNum, err := strconv.Atoi(string(hourSplitted[0]))
	if err != nil {
		return time.Time{}, fmt.Errorf("hour %w", err)
	}
	month, err := strconv.Atoi(string(dateSplitted[0]))
	if err != nil {
		return time.Time{}, fmt.Errorf("month %w", err)
	}
	day, err := strconv.Atoi(string(dateSplitted[1]))
	if err != nil {
		return time.Time{}, fmt.Errorf("day %w", err)
	}

	// TODO where can i take the year??
	t := time.Date(2024, time.Month(month), day, hourNum, 0, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	t = t.UTC()
	return t, nil
}
