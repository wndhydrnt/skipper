package args

import (
	"reflect"
	"testing"
	"time"

	"github.com/sanity-io/litter"
)

func TestNoArgsExpected(t *testing.T) {
	t.Run("fail, no args position", func(t *testing.T) {
		if err := Capture(); err == nil {
			t.Error("failed to fail")
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

	t.Run("pass, empty slice", func(t *testing.T) {
		if err := Capture(nil, []interface{}{}); err != nil {
			t.Error("failed", err)
		}
	})
}

func TestFixedArgs(t *testing.T) {
	var (
		oneInt      int
		oneFloat    float64
		oneString   string
		oneEnum     string
		oneDuration time.Duration

		expected = []interface{}{
			42,
			3.0,
			"foo",
			"red",
			3 * time.Second,
		}
	)

	run := func(title string, succeed bool, args []interface{}) {
		t.Run(title, func(t *testing.T) {
			oneInt = 0
			oneFloat = 0
			oneString = ""
			oneEnum = ""
			oneDuration = 0

			err := Capture(
				&oneInt,
				&oneFloat,
				&oneString,
				Enum(&oneEnum, "red", "green", "blue"),
				&oneDuration,
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
				oneDuration,
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
		"3s",
		"bar",
	})

	run("wrong type, not int", false, []interface{}{
		"not a number",
		3.0,
		"foo",
		"red",
		"3s",
	})

	run("wrong type, not float", false, []interface{}{
		42,
		"not a number",
		"foo",
		"red",
		"3s",
	})

	run("wrong type, not string", false, []interface{}{
		42,
		3.0,
		2,
		"red",
		"3s",
	})

	run("wrong enum, not string", false, []interface{}{
		42,
		3.0,
		"foo",
		42,
		"3s",
	})

	run("wrong enum", false, []interface{}{
		42,
		3.0,
		"foo",
		"cyan",
		"3s",
	})

	run("wrong type, not duration string", false, []interface{}{
		42,
		3.0,
		"foo",
		"red",
		"bar",
	})

	run("wrong type, not duration", false, []interface{}{
		42,
		3.0,
		"foo",
		"red",
		struct{}{},
	})

	run("pass", true, []interface{}{
		42,
		3.0,
		"foo",
		"red",
		"3s",
	})

	run("pass, convert int to float", true, []interface{}{
		42,
		3,
		"foo",
		"red",
		"3s",
	})

	run("pass, convert float to int", true, []interface{}{
		42.0,
		3.0,
		"foo",
		"red",
		"3s",
	})

	run("pass, duration as milliseconds", true, []interface{}{
		42,
		3.0,
		"foo",
		"red",
		3000,
	})

	run("pass, duration as float milliseconds", true, []interface{}{
		42,
		3.0,
		"foo",
		"red",
		3000.0,
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

		someInts      []int
		someFloats    []float64
		someStrings   []string
		someEnums     []string
		someDurations []time.Duration
		someMixed     []interface{}

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

		captureDurations = append(captureFixed, &someDurations)
		captureMixed     = append(captureFixed, &someMixed)

		argsFixed = []interface{}{
			42,
			"foo",
		}

		argsInts    = append(argsFixed, 1, 2, 3)
		argsFloats  = append(argsFixed, 1.41, 2.71, 3.14)
		argsStrings = append(argsFixed, "foo", "bar", "baz")
		argsEnums   = append(argsFixed, "green", "blue", "red")

		argsDurations = append(
			argsFixed,
			time.Hour.String(),
			time.Minute.String(),
			time.Second.String(),
		)

		argsMixed = append(argsFixed, 42, 3.14, "foo")

		expectedInts      = append(argsFixed, argsInts[len(argsFixed):])
		expectedFloats    = append(argsFixed, argsFloats[len(argsFixed):])
		expectedStrings   = append(argsFixed, argsStrings[len(argsFixed):])
		expectedEnums     = append(argsFixed, argsEnums[len(argsFixed):])
		expectedDurations = append(argsFixed, argsDurations[len(argsFixed):])
		expectedMixed     = append(argsFixed, argsMixed[len(argsFixed):])
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
		"fail, not duration",
		false,
		captureDurations,
		append(argsFixed, "3s", "not duration"),
		nil,
	)

	run(
		"pass, duration",
		true,
		captureDurations,
		argsDurations,
		expectedDurations,
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

	t.Run("missing optional", func(t *testing.T) {
		var (
			a int
			b string
		)

		if err := Capture(&a, Optional(&b), []interface{}{42}); err != nil {
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

	t.Run("too many with optional", func(t *testing.T) {
		var (
			a int
			b string
		)

		if err := Capture(&a, Optional(&b), []interface{}{42, "foo", "bar"}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("optional leaves the default", func(t *testing.T) {
		var a int
		b := "default value"
		if err := Capture(&a, Optional(&b), []interface{}{42}); err != nil {
			t.Error(err)
			return
		}

		if b != "default value" {
			t.Error("failed to leave the default value")
		}
	})

	t.Run("optional overrides the default with empty", func(t *testing.T) {
		var a int
		b := "default value"
		if err := Capture(&a, Optional(&b), []interface{}{42, ""}); err != nil {
			t.Error(err)
			return
		}

		if b != "" {
			t.Error("failed to leave the default value")
		}
	})
}

func TestDuration(t *testing.T) {
	t.Run("duration with unit", func(t *testing.T) {
		var d time.Duration
		if err := Capture(Duration(&d, time.Second), []interface{}{3.14}); err != nil {
			t.Error(err)
			return
		}

		if d < time.Duration(3.139*float64(time.Second)) ||
			d > time.Duration(3.141*float64(time.Second)) {
			t.Error("failed to parse duration with unit")
		}
	})

	t.Run("normalize optional operator", func(t *testing.T) {
		var d time.Duration
		if err := Capture(Duration(Optional(&d), time.Second), nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("invalid duration definition", func(t *testing.T) {
		var d string
		if err := Capture(Duration(&d, time.Second), []interface{}{42}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("variadic duration with unit", func(t *testing.T) {
		var d []time.Duration
		if err := Capture(Duration(&d, time.Second), []interface{}{1, 2, 3}); err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(d, []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}) {
			t.Error("failed to parse variadic durations")
		}
	})
}

func TestTime(t *testing.T) {
	t.Run("capture time string", func(t *testing.T) {
		var tim time.Time
		if err := Capture(&tim, []interface{}{"3000-12-18T09:36:18+09:00"}); err != nil {
			t.Error(err)
			return
		}

		expected := time.Date(3000, 12, 18, 9, 36, 18, 0, time.FixedZone("+0900", 32400))
		if !tim.Equal(expected) {
			t.Error("failed to parse time")
			t.Log("got:     ", tim)
			t.Log("expected:", expected)
		}
	})

	t.Run("capture time int", func(t *testing.T) {
		var tim time.Time
		if err := Capture(&tim, []interface{}{1511277065}); err != nil {
			t.Error(err)
			return
		}

		expected := time.Unix(1511277065, 0)
		if !tim.Equal(expected) {
			t.Error("failed to parse time")
			t.Log("got:     ", tim)
			t.Log("expected:", expected)
		}
	})

	t.Run("capture time float", func(t *testing.T) {
		var tim time.Time
		if err := Capture(&tim, []interface{}{1511277065.42}); err != nil {
			t.Error(err)
			return
		}

		expectedLess := time.Unix(1511277065, 39*int64(time.Second/time.Nanosecond)/100)
		expectedMore := time.Unix(1511277065, 43*int64(time.Second/time.Nanosecond)/100)
		if tim.Before(expectedLess) || tim.After(expectedMore) {
			t.Error("failed to parse time")
			t.Log("got:     ", tim)
		}
	})

	t.Run("invalid time string", func(t *testing.T) {
		var tim time.Time
		if err := Capture(&tim, []interface{}{"foo"}); err != ErrInvalidArgs {
			t.Error("failed to fail", err, ErrInvalidArgs)
		}
	})

	t.Run("not supported argument type", func(t *testing.T) {
		var tim time.Time
		if err := Capture(&tim, []interface{}{true}); err != ErrInvalidArgs {
			t.Error("failed to fail", err, ErrInvalidArgs)
		}
	})

	t.Run("capture variadic time arguments", func(t *testing.T) {
		var times []time.Time

		now := time.Now()
		expected := []time.Time{
			time.Unix(now.Add(-time.Second).Unix(), 0),
			time.Unix(now.Unix(), 0),
			time.Unix(now.Add(time.Second).Unix(), 0),
		}

		if err := Capture(&times, []interface{}{
			int(expected[0].Unix()),
			int(expected[1].Unix()),
			int(expected[2].Unix()),
		}); err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(times[0], expected[0]) {
			t.Error("failed to capture variadic times")
			t.Log("got:     ", times[0])
			t.Log("expected:", expected[0])
		}
	})

	t.Run("capture variadic time arguments, fail", func(t *testing.T) {
		var times []time.Time
		if err := Capture(&times, []interface{}{"foo"}); err != ErrInvalidArgs {
			t.Error("failed to fail", err, ErrInvalidArgs)
		}
	})
}

func TestDefault(t *testing.T) {
	i := 42
	if err := Capture(Optional(&i), nil); err != nil {
		t.Error(err)
		return
	}

	if i != 42 {
		t.Error("didn't leave the default")
	}
}

func TestVarargAfterOptional(t *testing.T) {
	t.Run("wrong type for optional", func(t *testing.T) {
		var (
			opt     int
			varargs []string
		)

		if err := Capture(
			Optional(&opt),
			&varargs,
			[]interface{}{"foo", "bar"},
		); err != ErrInvalidArgs {
			t.Error("failed to fail", err, ErrInvalidArgs)
		}
	})

	t.Run("ok", func(t *testing.T) {
		var (
			opt     int
			varargs []string
		)

		if err := Capture(
			Optional(&opt),
			&varargs,
			[]interface{}{42, "foo", "bar"},
		); err != nil {
			t.Error(err)
		}

		if opt != 42 || !reflect.DeepEqual(varargs, []string{"foo", "bar"}) {
			t.Error("failed to capture varargs after optional")
		}
	})
}
