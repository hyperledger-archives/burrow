package cli

import (
	"flag"

	"github.com/stretchr/testify/require"

	"io/ioutil"
	"os"
	"testing"
)

func TestTheCpCase(t *testing.T) {
	app := App("cp", "")
	app.Spec = "SRC... DST"

	src := app.Strings(StringsArg{Name: "SRC", Value: nil, Desc: ""})
	dst := app.String(StringArg{Name: "DST", Value: "", Desc: ""})

	ex := false
	app.Action = func() {
		ex = true
	}
	app.Run([]string{"cp", "x", "y", "z"})

	require.Equal(t, []string{"x", "y"}, *src)
	require.Equal(t, "z", *dst)

	require.True(t, ex, "Exec wasn't called")
}

func TestImplicitSpec(t *testing.T) {
	app := App("test", "")
	x := app.Bool(BoolOpt{Name: "x", Value: false, Desc: ""})
	y := app.String(StringOpt{Name: "y", Value: "", Desc: ""})
	called := false
	app.Action = func() {
		called = true
	}
	app.ErrorHandling = flag.ContinueOnError

	err := app.Run([]string{"test", "-x", "-y", "hello"})

	require.Nil(t, err)
	require.True(t, *x)
	require.Equal(t, "hello", *y)

	require.True(t, called, "Exec wasn't called")
}

func TestAppWithBoolOption(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue bool
	}{
		{[]string{"app"}, false},
		{[]string{"app", "-o"}, true},
		{[]string{"app", "-o=true"}, true},
		{[]string{"app", "-o=false"}, false},

		{[]string{"app", "--option"}, true},
		{[]string{"app", "--option=true"}, true},
		{[]string{"app", "--option=false"}, false},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.ErrorHandling = flag.ContinueOnError
		opt := app.BoolOpt("o option", false, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithStringOption(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue string
	}{
		{[]string{"app"}, "default"},
		{[]string{"app", "-o", "user"}, "user"},
		{[]string{"app", "-o=user"}, "user"},

		{[]string{"app", "--option", "user"}, "user"},
		{[]string{"app", "--option=user"}, "user"},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.ErrorHandling = flag.ContinueOnError
		opt := app.StringOpt("o option", "default", "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithIntOption(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue int
	}{
		{[]string{"app"}, 3},
		{[]string{"app", "-o", "16"}, 16},
		{[]string{"app", "-o=16"}, 16},

		{[]string{"app", "--option", "16"}, 16},
		{[]string{"app", "--option=16"}, 16},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.ErrorHandling = flag.ContinueOnError
		opt := app.IntOpt("o option", 3, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithStringsOption(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue []string
	}{
		{[]string{"app"}, []string{"a", "b"}},
		{[]string{"app", "-o", "x"}, []string{"x"}},
		{[]string{"app", "-o", "x", "-o=y"}, []string{"x", "y"}},

		{[]string{"app", "--option", "x"}, []string{"x"}},
		{[]string{"app", "--option", "x", "--option=y"}, []string{"x", "y"}},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.ErrorHandling = flag.ContinueOnError
		opt := app.StringsOpt("o option", []string{"a", "b"}, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithIntsOption(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue []int
	}{
		{[]string{"app"}, []int{1, 2}},
		{[]string{"app", "-o", "10"}, []int{10}},
		{[]string{"app", "-o", "10", "-o=11"}, []int{10, 11}},

		{[]string{"app", "--option", "10"}, []int{10}},
		{[]string{"app", "--option", "10", "--option=11"}, []int{10, 11}},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.ErrorHandling = flag.ContinueOnError
		opt := app.IntsOpt("o option", []int{1, 2}, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithBoolArg(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue bool
	}{
		{[]string{"app"}, false},
		{[]string{"app", "true"}, true},
		{[]string{"app", "false"}, false},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.Spec = "[ARG]"
		app.ErrorHandling = flag.ContinueOnError
		opt := app.BoolArg("ARG", false, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithStringArg(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue string
	}{
		{[]string{"app"}, "default"},
		{[]string{"app", "user"}, "user"},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.Spec = "[ARG]"
		app.ErrorHandling = flag.ContinueOnError
		opt := app.StringArg("ARG", "default", "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithIntArg(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue int
	}{
		{[]string{"app"}, 3},
		{[]string{"app", "16"}, 16},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.Spec = "[ARG]"
		app.ErrorHandling = flag.ContinueOnError
		opt := app.IntArg("ARG", 3, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithStringsArg(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue []string
	}{
		{[]string{"app"}, []string{"a", "b"}},
		{[]string{"app", "x"}, []string{"x"}},
		{[]string{"app", "x", "y"}, []string{"x", "y"}},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.Spec = "[ARG...]"
		app.ErrorHandling = flag.ContinueOnError
		opt := app.StringsArg("ARG", []string{"a", "b"}, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func TestAppWithIntsArg(t *testing.T) {

	cases := []struct {
		args             []string
		expectedOptValue []int
	}{
		{[]string{"app"}, []int{1, 2}},
		{[]string{"app", "10"}, []int{10}},
		{[]string{"app", "10", "11"}, []int{10, 11}},
	}

	for _, cas := range cases {
		t.Logf("Testing %#v", cas.args)

		app := App("app", "")
		app.Spec = "[ARG...]"
		app.ErrorHandling = flag.ContinueOnError
		opt := app.IntsArg("ARG", []int{1, 2}, "")

		ex := false
		app.Action = func() {
			ex = true
			require.Equal(t, cas.expectedOptValue, *opt)
		}
		err := app.Run(cas.args)

		require.NoError(t, err)
		require.True(t, ex, "Exec wasn't called")
	}
}

func testHelpAndVersionWithOptionsEnd(flag string, t *testing.T) {
	t.Logf("Testing help/version with --: flag=%q", flag)
	defer suppressOutput()()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 0, &exitCalled)()

	app := App("x", "")
	app.Version("v version", "1.0")
	app.Spec = "CMD"

	cmd := app.String(StringArg{Name: "CMD", Value: "", Desc: ""})

	actionCalled := false
	app.Action = func() {
		actionCalled = true
		require.Equal(t, flag, *cmd)
	}

	app.Run([]string{"x", "--", flag})

	require.True(t, actionCalled, "action should have been called")
	require.False(t, exitCalled, "exit should not have been called")

}

func TestHelpAndVersionWithOptionsEnd(t *testing.T) {
	for _, flag := range []string{"-h", "--help", "-v", "--version"} {
		testHelpAndVersionWithOptionsEnd(flag, t)
	}
}

var genGolden = flag.Bool("g", false, "Generate golden file(s)")

func TestHelpMessage(t *testing.T) {
	var out, err string
	defer captureAndRestoreOutput(&out, &err)()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 0, &exitCalled)()

	app := App("app", "App Desc")
	app.Spec = "[-bdsuikqs] BOOL1 [STR1] INT3..."

	// Options
	app.Bool(BoolOpt{Name: "b bool1 u uuu", Value: false, EnvVar: "BOOL1", Desc: "Bool Option 1"})
	app.Bool(BoolOpt{Name: "bool2", Value: true, EnvVar: " ", Desc: "Bool Option 2"})
	app.Bool(BoolOpt{Name: "d", Value: true, EnvVar: "BOOL3", Desc: "Bool Option 3", HideValue: true})

	app.String(StringOpt{Name: "s str1", Value: "", EnvVar: "STR1", Desc: "String Option 1"})
	app.String(StringOpt{Name: "str2", Value: "a value", Desc: "String Option 2"})
	app.String(StringOpt{Name: "u", Value: "another value", EnvVar: "STR3", Desc: "String Option 3", HideValue: true})

	app.Int(IntOpt{Name: "i int1", Value: 0, EnvVar: "INT1 ALIAS_INT1"})
	app.Int(IntOpt{Name: "int2", Value: 1, EnvVar: "INT2", Desc: "Int Option 2"})
	app.Int(IntOpt{Name: "k", Value: 1, EnvVar: "INT3", Desc: "Int Option 3", HideValue: true})

	app.Strings(StringsOpt{Name: "x strs1", Value: nil, EnvVar: "STRS1", Desc: "Strings Option 1"})
	app.Strings(StringsOpt{Name: "strs2", Value: []string{"value1", "value2"}, EnvVar: "STRS2", Desc: "Strings Option 2"})
	app.Strings(StringsOpt{Name: "z", Value: []string{"another value"}, EnvVar: "STRS3", Desc: "Strings Option 3", HideValue: true})

	app.Ints(IntsOpt{Name: "q ints1", Value: nil, EnvVar: "INTS1", Desc: "Ints Option 1"})
	app.Ints(IntsOpt{Name: "ints2", Value: []int{1, 2, 3}, EnvVar: "INTS2", Desc: "Ints Option 2"})
	app.Ints(IntsOpt{Name: "s", Value: []int{1}, EnvVar: "INTS3", Desc: "Ints Option 3", HideValue: true})

	// Args
	app.Bool(BoolArg{Name: "BOOL1", Value: false, EnvVar: "BOOL1", Desc: "Bool Argument 1"})
	app.Bool(BoolArg{Name: "BOOL2", Value: true, Desc: "Bool Argument 2"})
	app.Bool(BoolArg{Name: "BOOL3", Value: true, EnvVar: "BOOL3", Desc: "Bool Argument 3", HideValue: true})

	app.String(StringArg{Name: "STR1", Value: "", EnvVar: "STR1", Desc: "String Argument 1"})
	app.String(StringArg{Name: "STR2", Value: "a value", EnvVar: "STR2", Desc: "String Argument 2"})
	app.String(StringArg{Name: "STR3", Value: "another value", EnvVar: "STR3", Desc: "String Argument 3", HideValue: true})

	app.Int(IntArg{Name: "INT1", Value: 0, EnvVar: "INT1", Desc: "Int Argument 1"})
	app.Int(IntArg{Name: "INT2", Value: 1, EnvVar: "INT2", Desc: "Int Argument 2"})
	app.Int(IntArg{Name: "INT3", Value: 1, EnvVar: "INT3", Desc: "Int Argument 3", HideValue: true})

	app.Strings(StringsArg{Name: "STRS1", Value: nil, EnvVar: "STRS1", Desc: "Strings Argument 1"})
	app.Strings(StringsArg{Name: "STRS2", Value: []string{"value1", "value2"}, EnvVar: "STRS2"})
	app.Strings(StringsArg{Name: "STRS3", Value: []string{"another value"}, EnvVar: "STRS3", Desc: "Strings Argument 3", HideValue: true})

	app.Ints(IntsArg{Name: "INTS1", Value: nil, EnvVar: "INTS1", Desc: "Ints Argument 1"})
	app.Ints(IntsArg{Name: "INTS2", Value: []int{1, 2, 3}, EnvVar: "INTS2", Desc: "Ints Argument 2"})
	app.Ints(IntsArg{Name: "INTS3", Value: []int{1}, EnvVar: "INTS3", Desc: "Ints Argument 3", HideValue: true})

	app.Action = func() {}

	app.Command("command1", "command1 description", func(cmd *Cmd) {})
	app.Command("command2", "command2 description", func(cmd *Cmd) {})
	app.Command("command3", "command3 description", func(cmd *Cmd) {})

	app.Run([]string{"app", "-h"})

	if *genGolden {
		ioutil.WriteFile("testdata/help-output.txt.golden", []byte(err), 0644)
	}

	expected, e := ioutil.ReadFile("testdata/help-output.txt")
	require.NoError(t, e, "Failed to read the expected help output from testdata/help-output.txt")

	require.Equal(t, expected, []byte(err))
}

func TestLongHelpMessage(t *testing.T) {
	var out, err string
	defer captureAndRestoreOutput(&out, &err)()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 0, &exitCalled)()

	app := App("app", "App Desc")
	app.LongDesc = "Longer App Desc"
	app.Spec = "[-o] ARG"

	app.String(StringOpt{Name: "o opt", Value: "", Desc: "Option"})
	app.String(StringArg{Name: "ARG", Value: "", Desc: "Argument"})

	app.Action = func() {}
	app.Run([]string{"app", "-h"})

	if *genGolden {
		ioutil.WriteFile("testdata/long-help-output.txt.golden", []byte(err), 0644)
	}

	expected, e := ioutil.ReadFile("testdata/long-help-output.txt")
	require.NoError(t, e, "Failed to read the expected help output from testdata/long-help-output.txt")

	require.Equal(t, expected, []byte(err))
}

func TestVersionShortcut(t *testing.T) {
	defer suppressOutput()()
	exitCalled := false
	defer exitShouldBeCalledWith(t, 0, &exitCalled)()

	app := App("cp", "")
	app.Version("v version", "cp 1.2.3")

	actionCalled := false
	app.Action = func() {
		actionCalled = true
	}

	app.Run([]string{"cp", "--version"})

	require.False(t, actionCalled, "action should not have been called")
	require.True(t, exitCalled, "exit should have been called")
}

func TestSubCommands(t *testing.T) {
	app := App("say", "")

	hi, bye := false, false

	app.Command("hi", "", func(cmd *Cmd) {
		cmd.Action = func() {
			hi = true
		}
	})

	app.Command("byte", "", func(cmd *Cmd) {
		cmd.Action = func() {
			bye = true
		}
	})

	app.Run([]string{"say", "hi"})
	require.True(t, hi, "hi should have been called")
	require.False(t, bye, "byte should NOT have been called")
}

func TestContinueOnError(t *testing.T) {
	defer exitShouldNotCalled(t)()
	defer suppressOutput()()

	app := App("say", "")
	app.String(StringOpt{Name: "f", Value: "", Desc: ""})
	app.Spec = "-f"
	app.ErrorHandling = flag.ContinueOnError
	called := false
	app.Action = func() {
		called = true
	}

	err := app.Run([]string{"say"})
	require.NotNil(t, err)
	require.False(t, called, "Exec should NOT have been called")
}

func TestContinueOnErrorWithHelpAndVersion(t *testing.T) {
	defer exitShouldNotCalled(t)()
	defer suppressOutput()()

	app := App("say", "")
	app.Version("v", "1.0")
	app.String(StringOpt{Name: "f", Value: "", Desc: ""})
	app.Spec = "-f"
	app.ErrorHandling = flag.ContinueOnError
	called := false
	app.Action = func() {
		called = true
	}

	{
		err := app.Run([]string{"say", "-h"})
		require.Nil(t, err)
		require.False(t, called, "Exec should NOT have been called")
	}

	{
		err := app.Run([]string{"say", "-v"})
		require.Nil(t, err)
		require.False(t, called, "Exec should NOT have been called")
	}

}

func TestExitOnError(t *testing.T) {
	defer suppressOutput()()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 2, &exitCalled)()

	app := App("x", "")
	app.ErrorHandling = flag.ExitOnError
	app.Spec = "Y"

	app.String(StringArg{Name: "Y", Value: "", Desc: ""})

	app.Run([]string{"x", "y", "z"})
	require.True(t, exitCalled, "exit should have been called")
}

func TestExitOnErrorWithHelp(t *testing.T) {
	defer suppressOutput()()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 0, &exitCalled)()

	app := App("x", "")
	app.Spec = "Y"
	app.ErrorHandling = flag.ExitOnError

	app.String(StringArg{Name: "Y", Value: "", Desc: ""})

	app.Run([]string{"x", "-h"})
	require.True(t, exitCalled, "exit should have been called")
}

func TestExitOnErrorWithVersion(t *testing.T) {
	defer suppressOutput()()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 0, &exitCalled)()

	app := App("x", "")
	app.Version("v", "1.0")
	app.Spec = "Y"
	app.ErrorHandling = flag.ExitOnError

	app.String(StringArg{Name: "Y", Value: "", Desc: ""})

	app.Run([]string{"x", "-v"})
	require.True(t, exitCalled, "exit should have been called")
}

func TestPanicOnError(t *testing.T) {
	defer suppressOutput()()

	app := App("say", "")
	app.String(StringOpt{Name: "f", Value: "", Desc: ""})
	app.Spec = "-f"
	app.ErrorHandling = flag.PanicOnError
	called := false
	app.Action = func() {
		called = true
	}

	defer func() {
		if r := recover(); r != nil {
			require.False(t, called, "Exec should NOT have been called")
		} else {

		}
	}()
	app.Run([]string{"say"})
	t.Fatalf("wanted panic")
}

func TestOptSetByUser(t *testing.T) {
	cases := []struct {
		desc     string
		config   func(*Cli, *bool)
		args     []string
		expected bool
	}{
		// OPTS
		// String
		{
			desc: "String Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.String(StringOpt{Name: "f", Value: "a", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "value")
				c.String(StringOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.String(StringOpt{Name: "f", Value: "a", SetByUser: s})
			},
			args:     []string{"test", "-f=hello"},
			expected: true,
		},

		// Bool
		{
			desc: "Bool Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Bool(BoolOpt{Name: "f", Value: true, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "true")
				c.Bool(BoolOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Bool(BoolOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f"},
			expected: true,
		},

		// Int
		{
			desc: "Int Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Int(IntOpt{Name: "f", Value: 42, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "33")
				c.Int(IntOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Int(IntOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f=666"},
			expected: true,
		},

		// Ints
		{
			desc: "Ints Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Ints(IntsOpt{Name: "f", Value: []int{42}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "11,22,33")
				c.Ints(IntsOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Ints(IntsOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f=666"},
			expected: true,
		},

		// Strings
		{
			desc: "Strings Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Strings(StringsOpt{Name: "f", Value: []string{"aaa"}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "a,b,c")
				c.Strings(StringsOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Strings(StringsOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f=ccc"},
			expected: true,
		},
	}

	for _, cas := range cases {
		t.Log(cas.desc)

		setByUser := false
		app := App("test", "")

		cas.config(app, &setByUser)

		called := false
		app.Action = func() {
			called = true
		}

		app.Run(cas.args)

		require.True(t, called, "action should have been called")
		require.Equal(t, cas.expected, setByUser)
	}

}

func TestArgSetByUser(t *testing.T) {
	cases := []struct {
		desc     string
		config   func(*Cli, *bool)
		args     []string
		expected bool
	}{
		// ARGS
		// String
		{
			desc: "String Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.String(StringArg{Name: "ARG", Value: "a", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				os.Setenv("MOW_VALUE", "value")
				c.String(StringArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.String(StringArg{Name: "ARG", Value: "a", SetByUser: s})
			},
			args:     []string{"test", "aaa"},
			expected: true,
		},

		// Bool
		{
			desc: "Bool Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Bool(BoolArg{Name: "ARG", Value: true, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				os.Setenv("MOW_VALUE", "true")
				c.Bool(BoolArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Bool(BoolArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "true"},
			expected: true,
		},

		// Int
		{
			desc: "Int Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Int(IntArg{Name: "ARG", Value: 42, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				os.Setenv("MOW_VALUE", "33")
				c.Int(IntArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Int(IntArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "666"},
			expected: true,
		},

		// Ints
		{
			desc: "Ints Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Ints(IntsArg{Name: "ARG", Value: []int{42}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				os.Setenv("MOW_VALUE", "11,22,33")
				c.Ints(IntsArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Ints(IntsArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "333", "666"},
			expected: true,
		},

		// Strings
		{
			desc: "Strings Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Strings(StringsArg{Name: "ARG", Value: []string{"aaa"}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				os.Setenv("MOW_VALUE", "a,b,c")
				c.Strings(StringsArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Strings(StringsArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "aaa", "ccc"},
			expected: true,
		},
	}

	for _, cas := range cases {
		t.Log(cas.desc)

		setByUser := false
		app := App("test", "")

		cas.config(app, &setByUser)

		called := false
		app.Action = func() {
			called = true
		}

		app.Run(cas.args)

		require.True(t, called, "action should have been called")
		require.Equal(t, cas.expected, setByUser)
	}

}

func TestCommandAliases(t *testing.T) {
	defer suppressOutput()()

	cases := []struct {
		args          []string
		errorExpected bool
	}{
		{
			args:          []string{"say", "hello"},
			errorExpected: false,
		},
		{
			args:          []string{"say", "hi"},
			errorExpected: false,
		},
		{
			args:          []string{"say", "hello hi"},
			errorExpected: true,
		},
		{
			args:          []string{"say", "hello", "hi"},
			errorExpected: true,
		},
	}

	for _, cas := range cases {
		app := App("say", "")
		app.ErrorHandling = flag.ContinueOnError

		called := false

		app.Command("hello hi", "", func(cmd *Cmd) {
			cmd.Action = func() {
				called = true
			}
		})

		err := app.Run(cas.args)

		if cas.errorExpected {
			require.Error(t, err, "Run() should have returned with an error")
			require.False(t, called, "action should not have been called")
		} else {
			require.NoError(t, err, "Run() should have returned without an error")
			require.True(t, called, "action should have been called")
		}
	}
}

func TestSubcommandAliases(t *testing.T) {
	cases := []struct {
		args []string
	}{
		{
			args: []string{"app", "foo", "bar", "baz"},
		},
		{
			args: []string{"app", "foo", "bar", "z"},
		},
		{
			args: []string{"app", "foo", "b", "baz"},
		},
		{
			args: []string{"app", "f", "bar", "baz"},
		},
		{
			args: []string{"app", "f", "b", "baz"},
		},
		{
			args: []string{"app", "f", "b", "z"},
		},
		{
			args: []string{"app", "foo", "b", "z"},
		},
		{
			args: []string{"app", "f", "bar", "z"},
		},
	}

	for _, cas := range cases {
		app := App("app", "")
		app.ErrorHandling = flag.ContinueOnError

		called := false

		app.Command("foo f", "", func(cmd *Cmd) {
			cmd.Command("bar b", "", func(cmd *Cmd) {
				cmd.Command("baz z", "", func(cmd *Cmd) {
					cmd.Action = func() {
						called = true
					}
				})
			})
		})

		err := app.Run(cas.args)

		require.NoError(t, err, "Run() should have returned without an error")
		require.True(t, called, "action should have been called")
	}
}
