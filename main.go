package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"slices"

	"github.com/ross96D/cwbattle_parser/parser"
)

func main() {
	f1, err := os.Create("default.pprof")
	if err != nil {
		panic(err)
	}
	err = pprof.StartCPUProfile(f1)
	if err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	path := "battle_log.log"
	if len(os.Args) == 2 {
		path = os.Args[1]
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
	println("Battle resume by player")
	for _, k := range sort(m) {
		p := m[k]
		println(k.Name, "\tdone:", p.Damage, "\trecieved:", p.Tanqued, "\thits/total", fmt.Sprintf("%d/%d", p.Hits, p.Hits+p.Miss), fmt.Sprintf("%f%%", float64(p.Hits)/float64(p.Miss+p.Hits)))
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
