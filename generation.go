// That's "generation" as in a cohort of a particular age, not as in "the act of generating stuff".

package main

type Generation struct {
	Id int
}

const SCRIPTS_PER_GENERATION = 1000

func NewGeneration(id int) Generation {
	return Generation{Id: id}
}

// Ensure that we have a minimum number of scripts in the scripts folder.
func (g Generation) Initialize() {
	count := len(fileManager.ScriptIds)
	if count < SCRIPTS_PER_GENERATION {
		logger.Printf("Generating %d new scripts for generation %d", SCRIPTS_PER_GENERATION - count, g.Id)
		g.CreateScripts(SCRIPTS_PER_GENERATION - count)
	}
}

func (g Generation) CreateScripts(n int) {
	// FIXME: If a previous generation exists, splice and/or mutate their code

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
