package parser

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ross96D/battle-log-parser/assert"
)

var yellowRune, _ = utf8.DecodeRuneInString("🇻🇦")
var greenRune, _ = utf8.DecodeRuneInString("🇲🇴")
var blueRune, _ = utf8.DecodeRuneInString("🇪🇺")
var redRune, _ = utf8.DecodeRuneInString("🇮🇲")
var monsterRuneOnTurn, _ = utf8.DecodeRuneInString("⚱️")
var monsterRuneOnResume, _ = utf8.DecodeRuneInString("👹")

type Team byte

func (t Team) Name() string {
	switch t {
	case 'G':
		return "Green"
	case 'Y':
		return "Yellow"
	case 'B':
		return "Blue"
	case 'R':
		return "Red"
	case 'M':
		return "Monster"
	case 0:
		return "Miss"
	default:
		panic("unknow team " + string(t))
	}
}

func (t Team) String() string {
	switch t {
	case 'G':
		return "🇲🇴"
	case 'Y':
		return "🇻🇦"
	case 'B':
		return "🇪🇺"
	case 'R':
		return "🇮🇲"
	case 'M':
		return "👹"
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
	case "Green", "🇲🇴":
		*t = 'G'
	case "Yellow", "🇻🇦":
		*t = 'Y'
	case "Red", "🇮🇲":
		*t = 'R'
	case "Blue", "🇪🇺":
		*t = 'B'
	case "Monster", "👹":
		*t = 'M'
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
	case blueRune:
		return 'B'
	case redRune:
		return 'R'
	case monsterRuneOnResume, monsterRuneOnTurn:
		return 'M'
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
	if r == monsterRuneOnTurn {
		u.Name = s[size+3:]
	} else {
		u.Name = s[size+4:]
	}
	u.Name = strings.TrimSpace(u.Name)
	u.Team = TeamFromRune(r)
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

type Strike struct {
	Damage        int  `json:"damage"`
	TargetDefense int  `json:"target_defense"`
	Crit          bool `json:"crit"`
	Weakness      bool `json:"weakness"`
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

type Battle struct {
	Resume Resume    `json:"resume"`
	Turns  []Turn    `json:"turns"`
	Date   time.Time `json:"date"`
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
