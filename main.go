package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ross96D/battle-log-parser/parser"
	"github.com/ross96D/battle-log-parser/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var path string
var pprofPath string

var port uint16

func init() {
	rootCommand.AddCommand(&cliCommand)
	rootCommand.AddCommand(&serveCommand)

	cliCommand.Flags().StringVarP(&path, "input", "i", "battle_log.log", "the html file of the battle log to read")
	cliCommand.Flags().StringVar(&pprofPath, "pprof", "", "pprof file")

	serveCommand.Flags().Uint16VarP(&port, "port", "p", 0, "set the port to listen on")
	if err := serveCommand.MarkFlagRequired("port"); err != nil {
		panic(err)
	}
}

var rootCommand = cobra.Command{}

var cliCommand = cobra.Command{
	Use: "cli",
	Run: func(cmd *cobra.Command, args []string) {
		if pprofPath != "" {
			f1, err := os.Create("default.pprof")
			if err != nil {
				panic(err)
			}
			err = pprof.StartCPUProfile(f1)
			if err != nil {
				panic(err)
			}
			defer pprof.StopCPUProfile()
		}

		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		battle, err := parser.Parse(f)
		if err != nil {
			panic(err)
		}

		m := PlayerResumen(battle)
		// TODO print on the stdout instead of the stderr
		println("Battle resume by player")
		for _, k := range sort(m) {
			p := m[k]
			// println(k.Team.String(), "\t"+k.Name, "\tdone:", p.Damage, "\trecieved:", p.Tanqued, "\thits/total", fmt.Sprintf("%d/%d", p.Hits, p.Hits+p.Miss), fmt.Sprintf("%f%%", float64(p.Hits)/float64(p.Miss+p.Hits)), "\tCrits:", p.Crits)
			println(p.String())
		}
		println(battle.Date.String(), battle.Date.UnixMilli())
	},
}

var serveCommand = cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msgf("Serving on %d", port)
		if err := http.ListenAndServe(":"+strconv.FormatUint(uint64(port), 10), server.Server()); err != nil {
			log.Panic().Err(err).Send()
		}
	},
}

func main() {
	log.Logger = log.Output(zerolog.NewConsoleWriter())

	if err := rootCommand.Execute(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func sort(m map[parser.User]PlayerResume) []parser.User {
	type Z struct {
		k parser.User
		v PlayerResume
	}
	result := make([]Z, 0)
	for k, v := range m {
		result = append(result, Z{k: k, v: v})
	}
	slices.SortFunc(result, func(a Z, b Z) int {
		return b.v.Damage - a.v.Damage
	})
	ss := make([]parser.User, 0, len(result))
	for _, v := range result {
		ss = append(ss, v.k)
	}
	return ss
}

type PlayerResume struct {
	Team    parser.Team
	Damage  int
	Tanqued int
	Miss    int
	Hits    int
	Crits   int
	Name    string
}

func (pr PlayerResume) String() string {
	b := strings.Builder{}
	b.WriteString(pr.Team.String())
	b.WriteString(fmt.Sprintf(
		" %s\tdmg: %s\trecieved: %d\tHits/Total: %d/%d %.1f%%\tcrits: %d",
		pr.NameWithFixedWidth(13), FixedLenStr(strconv.FormatInt(int64(pr.Damage), 10), 5), pr.Tanqued, pr.Hits, pr.Hits+pr.Miss, 100*float64(pr.Hits)/float64(pr.Hits+pr.Miss), pr.Crits),
	)
	return b.String()
}

func (pr PlayerResume) NameWithFixedWidth(width uint) string {
	return FixedLenStr(pr.Name, width)
}

func (pr PlayerResume) Add(other PlayerResume) PlayerResume {
	return PlayerResume{
		Team:    pr.Team,
		Name:    pr.Name,
		Damage:  pr.Damage + other.Damage,
		Tanqued: pr.Tanqued + other.Tanqued,
		Hits:    pr.Hits + other.Hits,
		Miss:    pr.Miss + other.Miss,
		Crits:   pr.Crits + other.Crits,
	}
}

func PlayerResumen(b parser.Battle) map[parser.User]PlayerResume {
	empty := PlayerResume{}

	result := make(map[parser.User]PlayerResume, 0)
	for _, turn := range b.Turns {
		r := result[turn.Attacker]
		new := PlayerResume{
			Damage: turn.Damage(),
			Miss:   turn.Misses(),
			Hits:   turn.Hits(),
			Crits:  turn.Crits(),
		}
		if r == empty {
			r.Name = turn.Attacker.Name
			r.Team = turn.Attacker.Team
		}

		result[turn.Attacker] = r.Add(new)

		if !turn.Target.IsMiss() {
			r = result[turn.Target]
			if r == empty {
				r.Name = turn.Target.Name
				r.Team = turn.Target.Team
			}
			result[turn.Target] = r.Add(PlayerResume{Tanqued: turn.Damage(), Team: turn.Target.Team})
		}
	}
	return result
}

func FixedLenStr(str string, width uint) string {
	strB := []byte(str)
	nameLen := utf8.RuneCount(strB)
	if nameLen > int(width) {
		result := make([]byte, 0, width)
		for i, r := range str {
			if i == int(width) {
				break
			}
			if r < 128 {
				result = utf8.AppendRune(result, r)
			} else {
				_, size := utf8.DecodeRuneInString(string(r))
				if len(result)+size > int(width) {
					break
				}
				result = utf8.AppendRune(result, r)
			}
		}
		return string(result)
	} else {
		for i := uint(0); i < (width - uint(nameLen)); i++ {
			strB = append(strB, ' ')
		}
		return string(strB)
	}
}
