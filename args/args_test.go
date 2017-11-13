package args

import (
	"reflect"
	"testing"

	"github.com/sanity-io/litter"
)

func TestNoArgsExpected(t *testing.T) {
	t.Run("fail, no args position", func(t *testing.T) {
		if err := Capture(); err != ErrInvalidArgs {
			t.Error("failed to fail", err, ErrInvalidArgs)
		}
	})

	t.Run("fail, wrong args position", func(t *testing.T) {
		if err := Capture(struct{}{}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("fail", func(t *testing.T) {
		if err := Capture([]interface{}{"foo"}); err != ErrInvalidArgs {
			t.Error("failed to fail", err, ErrInvalidArgs)
		}
	})

	t.Run("pass", func(t *testing.T) {
		if err := Capture(nil); err != nil {
			t.Error("failed", err)
		}
	})
}

func TestFixedArgs(t *testing.T) {
	var (
		oneInt    int
		oneFloat  float64
		oneString string
		oneEnum   string

		expected = []interface{}{
			42,
			3.0,
			"foo",
			"red",
		}
	)

	run := func(title string, succeed bool, args []interface{}) {
		t.Run(title, func(t *testing.T) {
			oneInt = 0
			oneFloat = 0
			oneString = ""
			oneEnum = ""

			err := Capture(
				&oneInt,
				&oneFloat,
				&oneString,
				Enum(&oneEnum, "red", "green", "blue"),
				args,
			)

			if !succeed && err != ErrInvalidArgs {
				t.Error("failed to fail", err, ErrInvalidArgs)
				return
			}

			if succeed && err != nil {
				t.Error("failed", err)
				return
			}

			if !succeed {
				return
			}

			got := []interface{}{
				oneInt,
				oneFloat,
				oneString,
				oneEnum,
			}

			if !reflect.DeepEqual(got, expected) {
				t.Error("got wrong args", got, expected)
				t.Log("got:     ", litter.Sdump(got))
				t.Log("expected:", litter.Sdump(expected))
			}
		})
	}

	run("not enough", false, []interface{}{
		42,
		3.0,
		"foo",
	})

	run("too many", false, []interface{}{
		42,
		3.0,
		"foo",
		"red",
		"bar",
	})

	run("wrong type, not int", false, []interface{}{
		"not a number",
		3.0,
		"foo",
		"red",
	})

	run("wrong type, not float", false, []interface{}{
		42,
		"not a number",
		"foo",
		"red",
	})

	run("wrong type, not string", false, []interface{}{
		42,
		3.0,
		2,
		"red",
	})

	run("wrong enum, not string", false, []interface{}{
		42,
		3.0,
		"foo",
		42,
	})

	run("wrong enum", false, []interface{}{
		42,
		3.0,
		"foo",
		"cyan",
	})

	run("pass", true, []interface{}{
		42,
		3.0,
		"foo",
		"red",
	})

	run("pass, convert int to foat", true, []interface{}{
		42,
		3,
		"foo",
		"red",
	})

	run("pass, convert float to int", true, []interface{}{
		42.0,
		3.0,
		"foo",
		"red",
	})
}

func TestInvalidEnum(t *testing.T) {
	var v int
	if err := Capture(Enum(&v), []interface{}{42}); err == nil {
		t.Error("failed to fail")
	}
}

func TestNotSupportedType(t *testing.T) {
	var v int64
	if err := Capture(&v, []interface{}{int64(42)}); err == nil {
		t.Error("failed to fail")
	}
}

func TestOnlyVariadicArgs(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		var a []interface{}
		if err := Capture(&a, nil); err != nil {
			t.Error("failed", err)
		}
	})

	t.Run("some args", func(t *testing.T) {
		var a []interface{}
		if err := Capture(&a, []interface{}{"foo", "bar", "baz"}); err != nil {
			t.Error("failed", err)
		}

		if !reflect.DeepEqual(a, []interface{}{"foo", "bar", "baz"}) {
			t.Error("failed")
			t.Log("got:     ", litter.Sdump(a))
			t.Log("expected:", litter.Sdump([]interface{}{"foo", "bar", "baz"}))
		}
	})
}

func TestVariadicInWrongPosition(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		var (
			a, b int
			v    []int
		)

		if err := Capture(&a, &v, &b, []interface{}{1, 2, 3, 4}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("enums", func(t *testing.T) {
		var (
			a, b int
			v    []string
		)

		if err := Capture(&a, Enum(&v), &b, []interface{}{1, 2, 3, 4}); err == nil {
			t.Error("failed to fail")
		}
	})
}

func TestVariadicArgs(t *testing.T) {
	var (
		oneInt    int
		oneString string

		someInts    []int
		someFloats  []float64
		someStrings []string
		someEnums   []string
		someMixed   []interface{}

		captureFixed = []interface{}{
			&oneInt,
			&oneString,
		}

		captureInts    = append(captureFixed, &someInts)
		captureFloats  = append(captureFixed, &someFloats)
		captureStrings = append(captureFixed, &someStrings)

		captureEnums = append(
			captureFixed,
			Enum(
				&someEnums,
				"red",
				"green",
				"blue",
			),
		)

		captureMixed = append(captureFixed, &someMixed)

		argsFixed = []interface{}{
			42,
			"foo",
		}

		argsInts    = append(argsFixed, 1, 2, 3)
		argsFloats  = append(argsFixed, 1.41, 2.71, 3.14)
		argsStrings = append(argsFixed, "foo", "bar", "baz")
		argsEnums   = append(argsFixed, "green", "blue", "red")
		argsMixed   = append(argsFixed, 42, 3.14, "foo")

		expectedInts    = append(argsFixed, argsInts[len(argsFixed):])
		expectedFloats  = append(argsFixed, argsFloats[len(argsFixed):])
		expectedStrings = append(argsFixed, argsStrings[len(argsFixed):])
		expectedEnums   = append(argsFixed, argsEnums[len(argsFixed):])
		expectedMixed   = append(argsFixed, argsMixed[len(argsFixed):])
	)

	run := func(
		title string,
		succeed bool,
		capture []interface{},
		args []interface{},
		expected []interface{},
	) {
		t.Run(title, func(t *testing.T) {
			oneInt = 0
			oneString = ""
			someInts = nil
			someFloats = nil
			someStrings = nil
			someEnums = nil
			someMixed = nil

			err := Capture(append(capture, args)...)

			if !succeed && err != ErrInvalidArgs {
				t.Error("failed to fail", err, ErrInvalidArgs)
				return
			}

			if succeed && err != nil {
				t.Error("failed", err)
				return
			}

			if !succeed {
				return
			}

			fail := func() {
				t.Error("got wrong args", capture, expected)
				t.Log("got:     ", litter.Sdump(capture))
				t.Log("expected:", litter.Sdump(expected))
			}

			if len(capture) != len(expected) {
				fail()
				return
			}

			for i := range capture {
				switch v := capture[i].(type) {
				case *int:
					if *v != expected[i] {
						fail()
						return
					}
				case []int:
					exp, ok := expected[i].([]interface{})
					if !ok {
						fail()
						return
					}

					for j := range v {
						if v[i] != exp[j] {
							fail()
							return
						}
					}
				}
			}
		})
	}

	run(
		"less than fixed",
		false,
		captureInts,
		argsFixed[:1],
		nil,
	)

	var empty []int
	run(
		"pass, only fixed",
		true,
		captureInts,
		argsFixed,
		append(argsFixed, &empty),
	)

	run(
		"fail, not int",
		false,
		captureInts,
		append(argsFixed, 42, "not int"),
		nil,
	)

	run(
		"pass, ints",
		true,
		captureInts,
		argsInts,
		expectedInts,
	)

	run(
		"pass, convert floats to ints",
		true,
		captureInts,
		append(argsFixed, 1, 2.0, 3),
		expectedInts,
	)

	run(
		"fail, not float",
		false,
		captureFloats,
		append(argsFixed, 1.41, "not float"),
		nil,
	)

	run(
		"pass, floats",
		true,
		captureFloats,
		argsFloats,
		expectedFloats,
	)

	run(
		"pass, convert ints to floats",
		true,
		captureFloats,
		append(argsFixed, 1, 2.0, 3),
		expectedFloats,
	)

	run(
		"fail, not string",
		false,
		captureStrings,
		append(argsFixed, "foo", 42),
		nil,
	)

	run(
		"pass, strings",
		true,
		captureStrings,
		argsStrings,
		expectedStrings,
	)

	run(
		"fail, not enum",
		false,
		captureEnums,
		append(argsFixed, "red", 42),
		nil,
	)

	run(
		"fail, wrong enum",
		false,
		captureEnums,
		append(argsFixed, "red", "cyan"),
		nil,
	)

	run(
		"pass, enums",
		true,
		captureEnums,
		argsEnums,
		expectedEnums,
	)

	run(
		"fail, not supported mixed type",
		false,
		captureMixed,
		append(argsFixed, "foo", struct{}{}),
		nil,
	)

	run(
		"pass, mixed",
		true,
		captureMixed,
		argsMixed,
		expectedMixed,
	)
}

func TestOptionalArgs(t *testing.T) {
	t.Run("no optional", func(t *testing.T) {
		var (
			a int
			b string
		)

		if err := Capture(&a, &b, []interface{}{42, "foo"}); err != nil {
			t.Error(err)
		}
	})

	t.Run("non-optional after optional", func(t *testing.T) {
		var (
			a int
			b string
			c string
		)

		if err := Capture(&a, Optional(&b), Enum(&c, "true", "false"), []interface{}{
			42,
			"foo",
			"true",
		}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("optional args", func(t *testing.T) {
		var (
			a int
			b string
			c string
		)

		if err := Capture(&a, Optional(&b), Optional(&c), []interface{}{
			42,
			"foo",
			"bar",
		}); err != nil {
			t.Error(err)
			return
		}

		if a != 42 || b != "foo" || c != "bar" {
			t.Error("failed to capture args", a, b, c, 42, "foo", "bar")
		}
	})

	t.Run("optional arg as variadic", func(t *testing.T) {
		var (
			a int
			b []string
		)

		if err := Capture(&a, Optional(&b), []interface{}{42, "foo", "bar"}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("optional as enum", func(t *testing.T) {
		var (
			a int
			b string
		)

		if err := Capture(&a, Enum(Optional(&b), "true", "false"), []interface{}{
			42,
			"true",
		}); err != nil {
			t.Error(err)
			return
		}

		if a != 42 || b != "true" {
			t.Error("failed to capture args", a, b, 42, "true")
		}
	})

	t.Run("enum as optional", func(t *testing.T) {
		var (
			a int
			b string
		)

		if err := Capture(&a, Optional(Enum(&b, "true", "false")), []interface{}{
			42,
			"true",
		}); err != nil {
			t.Error(err)
			return
		}

		if a != 42 || b != "true" {
			t.Error("failed to capture args", a, b, 42, "true")
		}
	})
}
