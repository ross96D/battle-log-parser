package parser

import (
	"slices"
	"strings"

	"github.com/ross96D/battle-log-parser/assert"
	"golang.org/x/net/html"
)

func ParseResumeNode(n *html.Node) Resume {
	lines := getNodeLines(n)

	firstLine := lines[0]
	if strings.HasPrefix(firstLine, "📯Battle with") {
		return parseResumeNodeWithMonster(lines)
	}

	requiredFirstLine := "📯Battle for"

	assert.Assert(
		firstLine[0:len(requiredFirstLine)] == requiredFirstLine,
		"required first line to be equal to %s instead was %s",
		requiredFirstLine, firstLine[0:len(requiredFirstLine)],
	)
	assert.Assert(
		firstLine[len(firstLine)-1] == ']',
		"required first line to end on ] instead ended on %s",
		firstLine[len(firstLine)-1],
	)

	position := make([]byte, 0, 4)
	end := false
	for i := len(firstLine) - 1; i >= 0 && !end; i-- {
		char := firstLine[i]
		switch char {
		case ']', ' ':
			continue
		case '[':
			end = true
		default:
			position = append(position, char)
		}
	}
	slices.Reverse(position)

	resp := Resume{
		Position: NewMapPosition(position),
		Teams:    ParseResumeTeams(lines[2:]),
	}

	return resp
}

func parseResumeNodeWithMonster(lines []string) Resume {
	firstLine := lines[0]

	position := make([]byte, 0, 4)
	end := false
	for i := len(firstLine) - 1; i >= 0 && !end; i-- {
		char := firstLine[i]
		switch char {
		case ']', ' ':
			continue
		case '[':
			end = true
		default:
			position = append(position, char)
		}
	}
	slices.Reverse(position)

	resp := Resume{
		Position: NewMapPosition(position),
		Teams:    ParseResumeTeams(lines[2:]),
	}
	return resp
}

const greenPrefix = "🇲🇴Green Castle: "
const yellowPrefix = "🇻🇦Yellow Castle: "
const bluePrefix = "🇪🇺Blue Castle"
const redPrefix = "🇮🇲Red Castle"
const monsterPrefix = "👹Creatures"

func ParseResumeTeams(lines []string) []ResumeTeam {

	result := make([]ResumeTeam, 0)

	parse := func(line string) (total, alive uint64) {
		var err error
		i := strings.Index(line, " total") - 1
		assert.Assert(i != -1 && i != -2, "total not found in %s", line)
		total, err = parseNumBackwards(line, i)
		assert.NoError(err)

		i = strings.Index(line, " alive") - 1
		assert.Assert(i != -1 && i != -2, "total not found in %s", line)
		alive, err = parseNumBackwards(line, i)
		assert.NoError(err)

		return
	}

	for _, line := range lines {
		team := ResumeTeam{}
		if line, ok := strings.CutPrefix(line, greenPrefix); ok {
			total, alive := parse(line)
			team.Team = 'G'
			team.Alive = alive
			team.Total = total
		} else if line, ok := strings.CutPrefix(line, yellowPrefix); ok {
			total, alive := parse(line)
			team.Team = 'Y'
			team.Alive = alive
			team.Total = total
		} else if line, ok := strings.CutPrefix(line, bluePrefix); ok {
			total, alive := parse(line)
			team.Team = 'B'
			team.Alive = alive
			team.Total = total
		} else if line, ok := strings.CutPrefix(line, redPrefix); ok {
			total, alive := parse(line)
			team.Team = 'R'
			team.Alive = alive
			team.Total = total
		} else if line, ok := strings.CutPrefix(line, monsterPrefix); ok {
			total, alive := parse(line)
			team.Team = 'M'
			team.Alive = alive
			team.Total = total
		} else {
			break
		}
		result = append(result, team)
	}
	return result
}
