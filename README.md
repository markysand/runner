package runner // import "github.com/markysand/runner"

Package runner performs sequential running of tasks with log output for each
step. It is designed to run together with command line tools with an easily
configurable starting point.

VARIABLES

var (
	DoFormat   = "DO\t[%v/0-%v]\t%q"
	SkipFormat = "SKIP\t[%v/0-%v]\t%q"
)
    Default format strings for steps output


TYPES

type Step struct {
	Name      string       // The name of this step
	Run       func() error // The run function. A returned error will stop any subsequent runs
	Dependent bool         // A dependent step cannot be started from
}
    Step is a machine step for processing

type Steps []Step
    Steps is an array of Step

func (ss *Steps) Add(step Step) *Steps
    Add appends an additional runner step

func (ss Steps) GetStep(command string) (int, error)
    GetStep will return the starting index from a string command

func (ss Steps) Names() []string
    Names will get the names for the runner steps

func (ss Steps) Run(startIndex int) error
    Run from a specified zero based starting index

func (ss Steps) RunAll() error
    RunAll runs all steps in order

func (ss Steps) RunFromCommand(command string) error
    RunFromCommand will run from the specified command as name or as index
    number

