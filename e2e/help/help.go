// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

// This test sets singularity image specific environment variables and
// verifies that they are properly set.

package help

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sylabs/singularity/e2e/internal/e2e"
	"github.com/sylabs/singularity/internal/pkg/test/exec"
	"gotest.tools/assert"
	"gotest.tools/golden"
)

type ctx struct {
	env e2e.TestEnv
}

var helpContentTests = []struct {
	cmds []string
}{
	// singularity oci
	{[]string{"help", "oci"}},
	{[]string{"help", "oci", "attach"}},
	{[]string{"help", "oci", "create"}},
	{[]string{"help", "oci", "delete"}},
	{[]string{"help", "oci", "exec"}},
	{[]string{"help", "oci", "kill"}},
	{[]string{"help", "oci", "mount"}},
	{[]string{"help", "oci", "pause"}},
	{[]string{"help", "oci", "resume"}},
	{[]string{"help", "oci", "run"}},
	{[]string{"help", "oci", "start"}},
	{[]string{"help", "oci", "state"}},
	{[]string{"help", "oci", "umount"}},
	{[]string{"help", "oci", "update"}},
}

func (c *ctx) testHelpContent(t *testing.T) {
	for _, tc := range helpContentTests {
		name := fmt.Sprintf("%s.txt", strings.Join(tc.cmds, "-"))

		t.Run(name, func(t *testing.T) {
			path := filepath.Join("help", name)

			c := exec.Command(c.env.CmdPath, tc.cmds...)

			got := c.Run(t).Stdout()

			assert.Assert(t, golden.String(got, path))
		})
	}
}

func (c *ctx) testCommands(t *testing.T) {
	testCommands := []struct {
		name string
		argv []string
	}{
		{"Apps", []string{"apps"}},
		{"Bootstrap", []string{"bootstrap"}},
		{"Build", []string{"build"}},
		{"Check", []string{"check"}},
		{"Create", []string{"create"}},
		{"Exec", []string{"exec"}},
		{"Inspect", []string{"inspect"}},
		{"Mount", []string{"mount"}},
		{"Pull", []string{"pull"}},
		{"Run", []string{"run"}},
		{"Shell", []string{"shell"}},
		{"Test", []string{"test"}},
		{"InstanceDotStart", []string{"instance.start"}},
		{"InstanceDotList", []string{"instance.list"}},
		{"InstanceDotStop", []string{"instance.stop"}},
		{"InstanceStart", []string{"instance", "start"}},
		{"InstanceList", []string{"instance", "list"}},
		{"InstanceStop", []string{"instance", "stop"}},
	}

	for _, tt := range testCommands {

		testCmdsFn := func(t *testing.T, r *e2e.SingularityCmdResult) {

			testFlags := []struct {
				name string
				argv []string
				skip bool
			}{
				{"PostFlagShort", append(tt.argv, "-h"), true}, // TODO
				{"PostFlagLong", append(tt.argv, "--help"), false},
				{"PostCommand", append(tt.argv, "help"), false},
				{"PreFlagShort", append([]string{"-h"}, tt.argv...), false},
				{"PreFlagLong", append([]string{"--help"}, tt.argv...), false},
				{"PreCommand", append([]string{"help"}, tt.argv...), false},
			}

			for _, tf := range testFlags {
				if tf.skip && !c.env.RunDisabled {
					t.Skip("disabled until issue addressed")
				}

				e2e.RunSingularity(t, tf.name, e2e.WithArgs(tf.argv...),
					e2e.PostRun(func(t *testing.T) {
						if t.Failed() {
							t.Fatalf("Failed to run help flag while running command:\n%s\n", tt.name)
						}
					}),
					e2e.ExpectExit(0))
			}

		}

		e2e.RunSingularity(t, tt.name, e2e.WithArgs(tt.argv...),
			e2e.PostRun(func(t *testing.T) {
				if t.Failed() {
					t.Log("Failed to run help command")
				}
			}),
			e2e.ExpectExit(0, testCmdsFn))

	}

}

func (c *ctx) testFailure(t *testing.T) {
	if !c.env.RunDisabled {
		t.Skip("disabled until issue addressed") // TODO
	}

	tests := []struct {
		name string
		argv []string
	}{
		{"HelpBogus", []string{"help", "bogus"}},
		{"BogusHelp", []string{"bogus", "help"}},
		{"HelpInstanceBogus", []string{"help", "instance", "bogus"}},
		{"ImageBogusHelp", []string{"image", "bogus", "help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(c.env.CmdPath, tt.argv...)
			if res := cmd.Run(t); res.Error == nil {
				t.Fatalf("While running command:\n%s\nUnexpected success", res)
			}
		})
	}

}

func (c *ctx) testSingularity(t *testing.T) {
	tests := []struct {
		name       string
		argv       []string
		shouldPass bool
	}{
		{"NoCommand", []string{}, false},
		{"FlagShort", []string{"-h"}, true},
		{"FlagLong", []string{"--help"}, true},
		{"Command", []string{"help"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(c.env.CmdPath, tt.argv...)
			switch res := cmd.Run(t); {
			case res.Error == nil && tt.shouldPass:
				// expecting PASS, passed => PASS

			case res.Error != nil && !tt.shouldPass:
				// expecting FAIL, failed => PASS

			case res.Error == nil && !tt.shouldPass:
				// expecting PASS, failed => FAIL
				t.Fatalf("While running command:\n%s\nUnexpected failure: %+v",
					res,
					res.Error)

			case res.Error != nil && tt.shouldPass:
				// expecting FAIL, passed => FAIL
				t.Fatalf("While running command:\n%s\nUnexpected success", res)
			}
		})
	}

}

// RunE2ETests is the main func to trigger the test suite
func RunE2ETests(env e2e.TestEnv) func(*testing.T) {
	c := &ctx{
		env: env,
	}

	return func(t *testing.T) {
		// try to build from a non existen path
		t.Run("testCommands", c.testCommands)
		t.Run("testFailure", c.testFailure)
		t.Run("testSingularity", c.testSingularity)
		t.Run("testHelpContent", c.testHelpContent)
	}
}
