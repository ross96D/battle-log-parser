package parser

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ross96D/battle-log-parser/assert"
	"golang.org/x/net/html"
)

func ParseTurnNodes(nodes []*html.Node) []Turn {
	result := make([]Turn, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, ParseTurnNode(n))
	}
	return result
}

func ParseTurnNode(n *html.Node) Turn {
	lines := getNodeLines(n)
	attackerLine := lines[0]

	if len(lines) == 2 {
		return Turn{
			Attacker: parseAttackerLine(attackerLine),
		}
	}

	targetLine := lines[1]
	strikesLines := lines[2 : len(lines)-1]

	return Turn{
		Attacker: parseAttackerLine(attackerLine),
		Target:   parseTargeLine(targetLine),
		Strikes:  parseStrikesLines(strikesLines),
	}
}

func parseAttackerLine(line string) (u User) {
	assert.Assert(line != "")
	line, ok := strings.CutSuffix(line, " turn")
	assert.Assert(ok)
	return UserFromString(line)
}

func parseTargeLine(line string) User {
	assert.Assert(line != "")
	line, ok := strings.CutPrefix(line, "target: ")
	assert.Assert(ok)

	// get name string
	i := strings.Index(line, "HP, strikes: ")
	assert.Assert(i != -1)
	for ; i >= 0; i-- {
		if line[i] == ' ' {
			break
		}
	}

	return UserFromString(line[0:i])
}

func parseStrikesLines(lines []string) []Strike {
	if len(lines) == 0 {
		return []Strike{}
	}
	result := make([]Strike, 0, len(lines))
	for _, line := range lines {
		result = append(result, parseStrikeLine(line))
	}
	return result
}

var weaknessStrike, _ = utf8.DecodeRuneInString("⚡️")

func parseStrikeLine(line string) (strike Strike) {
	if line == "miss!" {
		return
	}
	r, size := utf8.DecodeRuneInString(line)
	if r == weaknessStrike {
		strike.Weakness = true
		line = line[size:]
		// it seems there is another character composing the ⚡️. Maybe is utf16 character
		_, size = utf8.DecodeRuneInString(line)
		if size >= 2 {
			line = line[size:]
		}
	}

	line, ok := strings.CutPrefix(line, "strike! dmg: ")
	if !ok {
		line, ok = strings.CutPrefix(line, "crit strike! dmg: ")
		strike.Crit = true
		assert.Assert(ok, line)
	}
	splitted := strings.Split(line, ". Pdef was: ")
	assert.Assert(len(splitted) == 2)

	dmg, err := strconv.Atoi(splitted[0])
	assert.NoError(err)
	defense, err := strconv.Atoi(splitted[1])
	assert.NoError(err)

	strike.Damage = dmg
	strike.TargetDefense = defense
	return
}
