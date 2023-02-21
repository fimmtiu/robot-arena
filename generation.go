// That's "generation" as in a cohort of a particular age, not as in "the act of generating stuff".

package main

import "math/rand"

type Generation struct {
	Id int
}

const SCRIPTS_PER_GENERATION = 1000
const RANDOM_PERCENT = 25
const MUTATE_PERCENT = 35
const SPLICE_PERCENT = 40

func NewGeneration(id int) Generation {
	return Generation{Id: id}
}

// Ensure that we have a minimum number of scripts in the scripts folder.
func (g Generation) Initialize() {
	count := len(fileManager.ScriptIds)
	if count < SCRIPTS_PER_GENERATION {
		logger.Printf("Generating %d new scripts for generation %d", SCRIPTS_PER_GENERATION - count, g.Id)
		n := rand.Intn(100)
		if n < RANDOM_PERCENT {
			g.MakeNewRandomScript()
		} else if n < RANDOM_PERCENT + MUTATE_PERCENT {
			// g.MutateScript FIXME which are the highest-rated scripts? Have to look that up first.
		}
		// g.CreateScripts(SCRIPTS_PER_GENERATION - count)
	}
}

func (g Generation) MutateScript(n int) {
	for i := 0; i < n; i++ {
		g.MakeNewRandomScript()
	}
}

func (g Generation) Run(arena *Arena) {

}

func (g Generation) MakeNewRandomScript() {
	code := ""

	file := fileManager.NewScriptFile()
	file.WriteString(code)
	file.Close()
}
