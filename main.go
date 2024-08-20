package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime/pprof"
	"slices"
	"strconv"

	"github.com/ross96D/cwbattle_parser/parser"
	"github.com/ross96D/cwbattle_parser/server"
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

		m := battle.PlayerResume()
		// TODO print on the stdout instead of the stderr
		println("Battle resume by player")
		for _, k := range sort(m) {
			p := m[k]
			println(k.Team.String(), "\t"+k.Name, "\tdone:", p.Damage, "\trecieved:", p.Tanqued, "\thits/total", fmt.Sprintf("%d/%d", p.Hits, p.Hits+p.Miss), fmt.Sprintf("%f%%", float64(p.Hits)/float64(p.Miss+p.Hits)), "\tCrits:", p.Crits)
		}
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

func sort(m map[parser.User]parser.PlayerResume) []parser.User {
	type Z struct {
		k parser.User
		v parser.PlayerResume
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
