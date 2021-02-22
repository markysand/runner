// Package runner performs sequential running of tasks
// with log output for each step.
// It is designed to run together with command line tools
// with an easily configurable starting point.
package runner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func getName(s Step, index int) string {
	return fmt.Sprintf("%v:%q", index, s.Name)
}

// Step is a machine step for processing
type Step struct {
	Name      string       // The name of this step
	Run       func() error // The run function. A returned error will stop any subsequent runs
	Dependent bool         // A dependent step cannot be started from
	SkipFunc               // Optional function to tell whether this step should be skipped
}

// SkipFunc is a function type to determine whether the step should be skipped
type SkipFunc func() bool

// SkipAlways is a SkipFunc that will have this step skipped always
var SkipAlways = SkipFunc(func() bool { return true })

func (skf SkipFunc) shouldSkip() bool {
	if skf == nil || !skf() {
		return false
	}
	return true
}

// Steps is an array of Step
type Steps []Step

// Names will get the names for the runner steps
func (ss Steps) Names() []string {
	result := make([]string, len(ss))
	for i := 0; i < len(ss); i++ {
		result[i] = getName(ss[i], i)
	}
	return result
}

// GetStep will return the starting index from a string command
func (ss Steps) GetStep(command string) (int, error) {
	if fromInt, err := strconv.Atoi(command); err == nil && fromInt >= 0 && fromInt < len(ss) {
		return fromInt, nil
	}

	for i, step := range ss {
		if step.Name == command {
			return i, nil
		}
	}

	return 0, errors.Errorf("%q is not a valid process step name, use: %v", command, strings.Join(ss.Names(), ", "))
}

// Run from a specified zero based starting index
func (ss Steps) Run(startIndex int) error {
	localLogger := log.Logger.With().Int("lastStep", len(ss)-1).Logger()

	if startStep := ss[startIndex]; startStep.Dependent {
		return errors.Errorf("step %v: %q cannot be started independently, it relies on previous steps", startIndex, startStep.Name)
	}
	for i, step := range ss {
		if i >= startIndex && !step.shouldSkip() {
			localLogger.Info().Int("step", i).Str("name", step.Name).Msg("do step")

			err := step.Run()
			if err != nil {
				return errors.Wrapf(err, "could not perform step %v, %v", i, step.Name)
			}
			continue
		}

		localLogger.Info().Int("step", i).Str("name", step.Name).Msg("skip step")
	}
	return nil
}

// RunAll runs all steps in order
func (ss Steps) RunAll() error {
	return ss.Run(0)
}

// RunFromCommand will run from the specified command
// as name or as index number
func (ss Steps) RunFromCommand(command string) error {
	start, err := ss.GetStep(command)

	if err != nil {
		return errors.Wrap(err, "could not parse command")
	}

	return ss.Run(start)
}

// Add appends an additional runner step
func (ss *Steps) Add(step Step) *Steps {
	*ss = append(*ss, step)
	return ss
}
