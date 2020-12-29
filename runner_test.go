// Package runner performs sequential running of tasks
// with log output for each step
package runner

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSteps_Names(t *testing.T) {
	tests := []struct {
		name string
		ss   Steps
		want []string
	}{
		{"standard flow",
			[]Step{
				{
					Name: "first",
				},
				{
					Name: "second",
				},
			},
			[]string{"0:\"first\"", "1:\"second\""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ss.Names(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Steps.Names() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSteps_GetStep(t *testing.T) {
	testSteps := []Step{
		{Name: "first"},
		{Name: "second"},
		{Name: "third"},
	}

	type args struct {
		command string
	}
	tests := []struct {
		name    string
		ss      Steps
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "name match",
			ss:   testSteps,
			args: args{
				command: "second",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "number match",
			ss:   testSteps,
			args: args{
				command: "2",
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "no match",
			ss:   testSteps,
			args: args{
				command: "foo",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "out of range",
			ss:   testSteps,
			args: args{
				command: "10",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ss.GetStep(tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("Steps.GetStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Steps.GetStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mock struct {
	called int
}

func (m *mock) success() error {
	m.called++
	return nil
}

func (m *mock) throw() error {
	m.called++
	return errors.New("error")
}

func TestSteps_Run(t *testing.T) {
	t.Run("standard flow, runs all steps, with nil error", func(t *testing.T) {
		f1, f2, f3 := new(mock), new(mock), new(mock)
		ss := Steps([]Step{{Run: f1.success}, {Run: f2.success}, {Run: f3.success}})

		err := ss.Run(0)

		assert.NoError(t, err)
		assert.Equal(t, 1, f1.called)
		assert.Equal(t, 1, f2.called)
		assert.Equal(t, 1, f3.called)
	})

	t.Run("error interrupt", func(t *testing.T) {
		f1, f2, f3 := new(mock), new(mock), new(mock)
		ss := Steps([]Step{{Run: f1.success}, {Run: f2.throw}, {Run: f3.success}})

		err := ss.Run(0)

		assert.Error(t, err)
		assert.Equal(t, 1, f1.called)
		assert.Equal(t, 1, f2.called)
		assert.Equal(t, 0, f3.called)
	})

	t.Run("starting on a Dependent step -> error", func(t *testing.T) {
		f1, f2, f3 := new(mock), new(mock), new(mock)
		ss := Steps([]Step{{Run: f1.success}, {Run: f2.success, Dependent: true}, {Run: f3.success}})

		err := ss.Run(1)

		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "cannot be started independently"), "wrong error type")
		assert.Equal(t, 0, f1.called)
		assert.Equal(t, 0, f2.called)
		assert.Equal(t, 0, f3.called)
	})

	t.Run("using start index", func(t *testing.T) {
		f1, f2, f3 := new(mock), new(mock), new(mock)
		ss := Steps([]Step{{Run: f1.success}, {Run: f2.success}, {Run: f3.success}})

		err := ss.Run(1)

		assert.NoError(t, err)
		assert.Equal(t, 0, f1.called)
		assert.Equal(t, 1, f2.called)
		assert.Equal(t, 1, f3.called)
	})
}

func TestSteps_RunAll(t *testing.T) {
	t.Run("standard flow, runs all steps, with nil error", func(t *testing.T) {
		f1, f2, f3 := new(mock), new(mock), new(mock)
		ss := Steps([]Step{{Run: f1.success}, {Run: f2.success}, {Run: f3.success}})

		err := ss.RunAll()

		assert.NoError(t, err)
		assert.Equal(t, 1, f1.called)
		assert.Equal(t, 1, f2.called)
		assert.Equal(t, 1, f3.called)
	})
}

func TestSteps_RunFromCommand(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		ss      Steps
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ss.RunFromCommand(tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("Steps.RunFromCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSteps_Add(t *testing.T) {
	before := Steps{{Name: "first"}}
	after := Steps{{Name: "first"}, {Name: "second"}}
	type args struct {
		step Step
	}
	tests := []struct {
		name string
		ss   *Steps
		args args
		want *Steps
	}{{
		name: "extending steps",
		ss:   &before,
		args: args{
			step: Step{
				Name:      "second",
				Run:       nil,
				Dependent: false,
			},
		},
		want: &after,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ss.Add(tt.args.step); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Steps.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
