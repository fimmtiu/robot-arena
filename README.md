# Robot Arena

A testbed for genetic algorithms where two teams of robots fight to win a laser tag-style game.

## Basic concepts

**Arena:** The robots are in a grid-based arena with walls that block line of sight and some weapons. The arena must be
horizontally and vertically symmetrical, because we want the scripts to execute identically regardless of which side of
the arena the team starts on. All

**Scoring:** Each team has a **goal** object on their side of the arena. If it's destroyed, they instantly lose. Scoring
is as follows:

* Killed a robot from the opposing team: +1 point
* Killed a robot from your own team: -2 points
* Game ends because an entire team is dead: +3 points to the winning team
* Game ends because it ran over time: -5 points to each team. This ensures that a stalemate game can never result in a positive score for either team.
* Game ends because the enemy's goal is destroyed: +10 points to the winners
* Game ends because you blew up your own goal: -20 points to the losers

The score can be used as the main fitness criterion.

**Robots:** Each team has five robots. They can move in orthogonal directions. They can shoot at each other with lasers
that are not 100% accurate; they get more accurate the longer the bot stands still, and less accurate the more often
they fire. The lasers have a cooldown if they get fired too often.

For now they'll have only one hit point and be omnidirectional turrets.

Idea for later: Add rockets. Each robot has one rocket. It explodes when it hits an obstacle, destroying all
non-boundary walls, robots, and goal objects within its radius. Destructible terrain would change things up in an
interesting way.

Possibilities: Hit points for robots, so they can take multiple laser hits? Facing, so they can only see and fire in a
certain arc in front of them?

**Time:** I had plans to make a more complicated real-time thing, but let's just make it turn-based for now. Each robot
can take one action (move, shoot, wait, etc.) per tick of the simulation. A game that hasn't completed after 2,000 ticks
ends automatically because the robots are probably just stuck in corners running into walls like idiots.

## Script components

We want to reduce the language down to a single type: integers. No booleans, coordinate pairs, floats, etc. As such, all
predicates will return 1 if true and 0 if false, and all conditionals will check for non-zero values.

All directions are relative to the team's starting orientation. The enemy's goal is north and your goal is south.

### Language primitives

`if`, `and`, `or`, `not`. Like the standard primitives, but they check if something is non-zero rather than true.

### Predicates

`north?`: True if the north cell is empty
`south?`: True if the south cell is empty
`east?`: True if the east cell is empty
`west?`: True if the west cell is empty
`can-fire?`: True if your laser isn't overheated
`enemy-visible?`: True if any enemy robot is within sight
`enemy-goal-visible?`: True if the enemy's goal is within sight
`own-goal-visible?`: True if your own goal is within sight

### Actions

All actions change the unit's accuracy penalty. Shooting increases the penalty, up to a certain maximum. Waiting or moving
lowers it closer to 0 (waiting moreso than moving). This is intended to make "static turret" behaviour less likely to
evolve.

Ideas: Base chance to hit is 100% - (target_distance * 5). Each shot adds +10% miss chance, up to +50%.
All games should be seeded with the run number.

`(go n)`: N represents a direction: 1 = north, 2 = east, 3 = south, 4 = west.
`(go-north)`: Move one cell to the north. Returns 1 if successful and 0 if you hit an obstacle.
`(go-south)`: Move one cell to the south. Returns 1 if successful and 0 if you hit an obstacle.
`(go-east)`: Move one cell to the east. Returns 1 if successful and 0 if you hit an obstacle.
`(go-west)`: Move one cell to the west. Returns 1 if successful and 0 if you hit an obstacle.
`(wait)`: Do nothing.
`(shoot-nearest)`: Fire a laser at the nearest enemy or goal. Returns 1 if successful and 0 if your laser is overheated or there's nobody in sight.
`(shoot-random)`: Fire a laser at a random enemy or goal. Returns 1 if successful and 0 if your laser is overheated or there's nobody in sight.

### Math

`(+ int int)`
`(- int int)`
`(* int int)`
`(/ int int)`  (note: integer division only.)
`(% int int)`
`>`
`<`
`=`



### Other functions

`(tick)`: The tick number. Starts at 0, goes up to 2,000 or so.
`(nearest-visible-enemy)`: The id of the nearest visible enemy robot in range
`(number-of-visible-enemies)`: The number of visible enemies
`(my-x-pos)`: The X coordinate of the current unit (rotated relative to the team's NSEW orientation)
`(my-y-pos)`: The Y coordinate of the current unit (rotated relative to the team's NSEW orientation)


## Input

Command line: the scenario name, the action to take.

Actions:

`run <scenario> <number of generations>`: runs the simulation for N generations
`run-once <scenario>`: Runs the given generation
`generate <scenario> <generation> <n>`: create N completely random new scripts
`mutate <scenario> <generation> <script id>`: mutate the given script in some small way
`splice <scenario> <generation> <id1> <id2>`: Cross-pollinate a random expression between two scripts in the same generation


## Output

On screen: Update terminal output once per second with:
  * Game number in this generation, out of total
  * Score for each team

When a game ends, terminal output for the final score.

Files:

`scenario/foo/gen_1/cells`: A file containing a series of these binary packed structures:
  * First two bits: type (move, shoot, kill)
  * 7 bits: x coord
  * 7 bits: y coord
  * 2 bytes: Number of events

Each complete game will emit one record of each type for each non-zero square on the board. (Only record successful
moves, not bumping-into-wall failures.)

`scenario/foo/gen_1/results`: A CSV file with the following header:

generation,matchId,scriptA,scriptB,scoreA,scoreB,ticks

`scenario/foo/gen_1/scripts/1.l`: A file of auto-generated RoboScript code.
