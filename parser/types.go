package parser

import (
	"bytes"
	"errors"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ross96D/cwbattle_parser/assert"
	"golang.org/x/net/html"
)

var yellowRune, _ = utf8.DecodeRuneInString("🇻🇦")
var greenRune, _ = utf8.DecodeRuneInString("🇲🇴")

type Team byte

func (t Team) String() string {
	switch t {
	case 'G':
		return "Green"
	case 'Y':
		return "Yellow"
	case 'B':
		return "Blue"
	case 'R':
		return "Red"
	case 0:
		return "Miss"
	default:
		panic("unknow team " + string(t))
	}
}

func (t Team) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

func (t *Team) UnmarshalJSON(b []byte) error {
	s := strings.TrimFunc(string(b), func(r rune) bool {
		return r == '"'
	})
	switch s {
	case "Green":
		*t = 'G'
	case "Yellow":
		*t = 'Y'
	case "Red":
		*t = 'R'
	case "Blue":
		*t = 'B'
	case "Miss":
		*t = 0
	default:
		return errors.New("invalid team " + string(b))
	}
	return nil
}

func TeamFromRune(r rune) Team {
	switch r {
	case yellowRune:
		return 'Y'
	case greenRune:
		return 'G'
	default:
		panic("unidentify rune " + string(r))
	}
}

type User struct {
	Team Team   `json:"team"`
	Name string `json:"name"`
}

func (u User) String() string {
	return u.Team.String() + " " + u.Name
}

func (u User) IsMiss() bool {
	return u == User{}
}

func UserFromString(s string) (u User) {
	// TODO flag icon is composed of 2 runes but i work as if it is one
	r, size := utf8.DecodeRune([]byte(s))
	assert.Assert(r != utf8.RuneError)
	u.Team = TeamFromRune(r)
	u.Name = s[size+4:]
	return u
}

type Position struct {
	Team Team
	Y    uint64
	X    uint64
}

func (p Position) String() string {
	b := strings.Builder{}
	b.WriteByte(byte(p.Team))
	b.WriteString(strconv.FormatUint(p.Y, 10))
	b.WriteByte('#')
	b.WriteString(strconv.FormatUint(p.X, 10))

	return b.String()
}

func NewMapPosition(position []byte) Position {
	pos := Position{}
	switch position[0] {
	case 'Y', 'G', 'R', 'B':
		pos.Team = Team(position[0])
	default:
		panic("invalid position string " + string(position))
	}
	coord := bytes.Split(position[1:], []byte("#"))
	if len(coord) == 1 {
		switch coord[0][0] {
		case 'Y', 'B':
			pos.Y = 0
			x, err := strconv.ParseUint(string(coord[0][1:]), 10, 64)
			assert.NoError(err, "line %s", string(position))
			pos.X = x
		case 'R', 'G':
			pos.X = 0
			y, err := strconv.ParseUint(string(coord[0][1:]), 10, 64)
			assert.NoError(err, "line %s", string(position))
			pos.X = y
		}

	} else {
		y, err := strconv.ParseUint(string(coord[0]), 10, 64)
		assert.NoError(err, "line %s", string(position))
		x, err := strconv.ParseUint(string(coord[1]), 10, 64)
		assert.NoError(err, "line %s", string(position))

		pos.Y = y
		pos.X = x
	}

	return pos
}

type ResumeTeam struct {
	Total uint64
	Alive uint64
	Team  byte
}

func (rt ResumeTeam) String() string {
	b := strings.Builder{}
	b.WriteByte(rt.Team)
	b.WriteByte(' ')
	b.WriteString("total: " + strconv.FormatUint(rt.Total, 10))
	b.WriteString(" alive: " + strconv.FormatUint(rt.Alive, 10))
	return b.String()
}

type Resume struct {
	Position Position
	Teams    []ResumeTeam
}

func (r Resume) String() string {
	b := strings.Builder{}
	b.WriteString(r.Position.String())
	b.WriteByte('\n')
	b.WriteString("teams:")

	for _, t := range r.Teams {
		b.WriteString("\t" + t.String())
		b.WriteByte('\n')
	}
	return b.String()
}

func ParseResumeNode(n *html.Node) Resume {
	lines := getNodeLines(n)

	firstLine := lines[0]
	requiredFirstLine := "📯Battle for"

	assert.Assert(
		firstLine[0:len(requiredFirstLine)] == requiredFirstLine,
		"required first line to be equal to %s instead was",
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

func ParseResumeTeams(lines []string) []ResumeTeam {
	const greenPrefix = "🇲🇴Green Castle: "
	const yellowPrefix = "🇻🇦Yellow Castle: "

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
		} else {
			break
		}
		result = append(result, team)
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

func ParseTurnNodes(nodes []*html.Node) []Turn {
	result := make([]Turn, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, ParseTurnNode(n))
	}
	return result
}

type Strike struct {
	Damage        int  `json:"damage"`
	TargetDefense int  `json:"target_defense"`
	Crit          bool `json:"crit"`
}

func (s Strike) String() string {
	empty := Strike{}
	if s == empty {
		return "Miss"
	}
	return "Damage: " + strconv.FormatInt(int64(s.Damage), 10) + " Def" + strconv.FormatInt(int64(s.TargetDefense), 10)
}

func (s Strike) IsMiss() bool {
	return s == Strike{}
}

type Turn struct {
	Attacker User     `json:"atacker"`
	Target   User     `json:"target"`
	Strikes  []Strike `json:"strikes"`
}

func (t Turn) Misses() int {
	count := 0
	for _, strike := range t.Strikes {
		if strike.IsMiss() {
			count++
		}
	}
	return count
}

func (t Turn) Hits() int {
	return len(t.Strikes) - t.Misses()
}

func (t Turn) Crits() int {
	count := 0
	for _, strike := range t.Strikes {
		if strike.Crit {
			count++
		}
	}
	return count
}

func (t Turn) Damage() int {
	result := 0
	for _, s := range t.Strikes {
		result += s.Damage
	}
	return result
}

func (t Turn) String() string {
	b := strings.Builder{}
	b.WriteString("Turn. Attacker: " + t.Attacker.String())
	b.WriteByte(' ')
	b.WriteString("Target: " + t.Target.String())
	b.WriteString("\nStrikes:")
	for _, strike := range t.Strikes {
		b.WriteString("\n\t")
		b.WriteString(strike.String())
	}
	return b.String()
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

func parseStrikeLine(line string) (strike Strike) {
	if line == "miss!" {
		return
	}
	line, ok := strings.CutPrefix(line, "strike! dmg: ")
	if !ok {
		line, ok = strings.CutPrefix(line, "crit strike! dmg: ")
		strike.Crit = true
		assert.Assert(ok)
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

type Battle struct {
	Resume Resume `json:"resume"`
	Turns  []Turn `json:"turns"`
}

func (b Battle) PlayerListWithDamage() map[User]int {
	result := make(map[User]int, 0)
	for _, turn := range b.Turns {
		r, ok := result[turn.Attacker]
		if !ok {
			result[turn.Attacker] = turn.Damage()
		} else {
			result[turn.Attacker] = r + turn.Damage()
		}
	}
	return result
}

type PlayerResume struct {
	Team    Team
	Damage  int
	Tanqued int
	Miss    int
	Hits    int
	Crits   int
}

func (pr PlayerResume) Add(other PlayerResume) PlayerResume {
	assert.Assert(pr.Team == Team(0) || pr.Team == other.Team)
	return PlayerResume{
		Team:    pr.Team,
		Damage:  pr.Damage + other.Damage,
		Tanqued: pr.Tanqued + other.Tanqued,
		Hits:    pr.Hits + other.Hits,
		Miss:    pr.Miss + other.Miss,
		Crits:   pr.Crits + other.Crits,
	}
}

func (b Battle) PlayerResume() map[User]PlayerResume {
	result := make(map[User]PlayerResume, 0)
	for _, turn := range b.Turns {
		r := result[turn.Attacker]
		new := PlayerResume{
			Team:   turn.Attacker.Team,
			Damage: turn.Damage(),
			Miss:   turn.Misses(),
			Hits:   turn.Hits(),
			Crits:  turn.Crits(),
		}

		result[turn.Attacker] = r.Add(new)

		if !turn.Target.IsMiss() {
			r = result[turn.Target]
			result[turn.Target] = r.Add(PlayerResume{Tanqued: turn.Damage(), Team: turn.Target.Team})
		}
	}
	return result
}
