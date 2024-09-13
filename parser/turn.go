package parser

import (
	"strconv"
	"strings"

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
	strikesLines := strikeLines(lines[2:])

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
		strike, ok := parseStrikeLine(line)
		if !ok {
			continue
		}
		result = append(result, strike)
	}
	return result
}

func parseStrikeLine(line string) (Strike, bool) {
	origLine := line

	line, modifier := stripSymbols(line)
	if modifier == counter {
		index := strings.Index(line, "strike! dmg:")
		// TODO miss on counter
		if index == -1 {
			return Strike{}, false
		}
		line = line[strings.Index(line, "strike! dmg:"):]
	}
	if line == "miss!" {
		return Strike{}, true
	}

	line, ok := cutPrefixStrikeLines(line)
	assert.Assert(ok, origLine)
	splitted := strings.Split(line, ". Pdef was: ")
	assert.Assert(len(splitted) == 2)

	dmg, err := strconv.Atoi(splitted[0])
	assert.NoError(err)
	defense, err := strconv.Atoi(splitted[1])
	assert.NoError(err)

	strike := Strike{}
	strike.Damage = dmg
	strike.TargetDefense = defense
	return strike, true
}

var symbols map[string]string = map[string]string{
	"weaknessStrike": "⚡️",
	"unkownWater":    "💦",
	"unkownCruz":     "➕",
	"counterAttack":  "🔄",
}

type attacksModifiers int

func (attacksModifiers) fromStr(s string) attacksModifiers {
	switch s {
	case "weaknessStrike":
		return weakness
	case "unkownWater":
		return unkown
	case "unkownCruz":
		return unkown
	case "counterAttack":
		return counter
	default:
		return unkown
	}
}

const (
	none attacksModifiers = iota - 1
	unkown
	weakness
	counter
)

// TODO add value
func stripSymbols(line string) (string, attacksModifiers) {
	for k, v := range symbols {
		l := len(v)
		if l >= len(line) {
			continue
		}
		prefix := line[0:l]
		if v == prefix {
			return line[l:], attacksModifiers(0).fromStr(k)
		}
	}
	return line, none
}

func cutPrefixStrikeLines(line string) (string, bool) {
	if line, ok := strings.CutPrefix(line, "strike! dmg: "); ok {
		return line, true
	}
	if line, ok := strings.CutPrefix(line, "crit strike! dmg: "); ok {
		return line, true
	}
	if line, ok := strings.CutPrefix(line, "💦strike! dmg: "); ok {
		return line, true
	}
	if line, ok := strings.CutPrefix(line, "💦crit strike! dmg: "); ok {
		return line, true
	}
	return line, false
}

func strikeLines(lines []string) []string {
	end := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if hasFlagAtBegining(line) {
			end--
			continue
		}
		if strings.HasSuffix(line, "retrieved an arrow") {
			end--
			continue
		}
		break
	}
	assert.Assert(end > 0, strconv.Itoa(end)+": "+strings.Join(lines, "\n"))
	return lines[:end]
}

func hasFlagAtBegining(line string) bool {
	_, _, err := TeamFromRune(line)
	return err == nil
}
