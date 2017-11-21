package args

import (
	"errors"
	"time"
)

type enum struct {
	options []string
	p       interface{}
	invalid bool
}

type duration struct {
	unit    time.Duration
	p       interface{}
	invalid bool
}

type optional struct {
	p interface{}
}

type captureFlags int

const (
	none    captureFlags = 0
	hasArgs captureFlags = 1 << iota
	expectingOptional
	expectingSingle
	lastPosition
	varargsConsumed
)

var (
	ErrInvalidArgs = errors.New("invalid args")

	errArgPositionType          = errors.New("last argument must be of type []interface{} or nil")
	errNotSupportedCaptureType  = errors.New("not supported capture type")
	errNotVariadicPosition      = errors.New("variadic must be the last arg")
	errInvalidEnum              = errors.New("invalid enum")
	errExpetingOptionalOrVararg = errors.New("expecting optional or vararg")
	errInvalidDurationCapture   = errors.New("invalid duration definition capture")
)

func splitArgs(a []interface{}) (captures []interface{}, args []interface{}, err error) {
	if len(a) == 0 {
		err = errArgPositionType
		return
	}

	last := len(a) - 1
	var lastItem interface{}
	captures, lastItem = a[:last], a[last]

	if lastItem == nil {
		return
	}

	var ok bool
	args, ok = lastItem.([]interface{})
	if !ok {
		err = errArgPositionType
	}

	return
}

func validateCapture(capture interface{}, f captureFlags) error {
	switch p := capture.(type) {
	case *int, *float64, *string, *time.Duration, *time.Time:
		if f&expectingOptional != 0 {
			return errExpetingOptionalOrVararg
		}

		if f&hasArgs == 0 {
			return ErrInvalidArgs
		}
	case duration:
		if p.invalid {
			return errInvalidDurationCapture
		}

		return validateCapture(p.p, f)
	case *[]int, *[]float64, *[]string, *[]time.Duration, *[]time.Time, *[]interface{}:
		if f&expectingSingle != 0 || f&lastPosition == 0 {
			return errNotVariadicPosition
		}
	case enum:
		if p.invalid {
			return errInvalidEnum
		}

		return validateCapture(p.p, f)
	case optional:
		if f&hasArgs == 0 {
			return nil
		}

		return validateCapture(p.p, f&^expectingOptional|expectingSingle)
	}

	return nil
}

func captureInt(a interface{}) (int, error) {
	switch v := a.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, ErrInvalidArgs
	}
}

func captureInts(a []interface{}) ([]int, error) {
	var ints []int
	for i := range a {
		v, err := captureInt(a[i])
		if err != nil {
			return nil, err
		}

		ints = append(ints, v)
	}

	return ints, nil
}

func captureFloat(a interface{}) (float64, error) {
	switch v := a.(type) {
	case int:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, ErrInvalidArgs
	}
}

func captureFloats(a []interface{}) ([]float64, error) {
	var floats []float64
	for i := range a {
		v, err := captureFloat(a[i])
		if err != nil {
			return nil, err
		}

		floats = append(floats, v)
	}

	return floats, nil
}

func captureString(a interface{}) (string, error) {
	v, ok := a.(string)
	if !ok {
		return "", ErrInvalidArgs
	}

	return v, nil
}

func captureStrings(a []interface{}) ([]string, error) {
	var strings []string
	for i := range a {
		v, err := captureString(a[i])
		if err != nil {
			return nil, err
		}

		strings = append(strings, v)
	}

	return strings, nil
}

func captureEnum(options []string, a interface{}) (string, error) {
	v, ok := a.(string)
	if !ok {
		return "", ErrInvalidArgs
	}

	for i := range options {
		if options[i] == v {
			return v, nil
		}
	}

	return "", ErrInvalidArgs
}

func captureEnums(options []string, a []interface{}) ([]string, error) {
	var enums []string
	for i := range a {
		v, err := captureEnum(options, a[i])
		if err != nil {
			return nil, err
		}

		enums = append(enums, v)
	}

	return enums, nil
}

func captureDuration(a interface{}, unit time.Duration) (time.Duration, error) {
	switch v := a.(type) {
	case int:
		return time.Duration(v) * unit, nil
	case float64:
		return time.Duration(v * float64(unit)), nil
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			return 0, ErrInvalidArgs
		}

		return d, nil
	default:
		return 0, ErrInvalidArgs
	}
}

func captureDurations(a []interface{}, unit time.Duration) ([]time.Duration, error) {
	var durations []time.Duration
	for i := range a {
		v, err := captureDuration(a[i], unit)
		if err != nil {
			return nil, ErrInvalidArgs
		}

		durations = append(durations, v)
	}

	return durations, nil
}

func captureTime(a interface{}) (time.Time, error) {
	switch v := a.(type) {
	case int:
		return time.Unix(int64(v), 0), nil
	case float64:
		return time.Unix(
			int64(v),
			int64((v - float64(int(v))) * float64(time.Second / time.Nanosecond)),
		), nil
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			err = ErrInvalidArgs
		}

		return t, err
	default:
		return time.Time{}, ErrInvalidArgs
	}
}

func captureTimes(a []interface{}) ([]time.Time, error) {
	var times []time.Time
	for i := range a {
		v, err := captureTime(a[i])
		if err != nil {
			return nil, ErrInvalidArgs
		}

		times = append(times, v)
	}

	return times, nil
}

func captureMixed(a []interface{}) ([]interface{}, error) {
	var mixed []interface{}
	for i := range a {
		switch a[i].(type) {
		case int, float64, string:
			mixed = append(mixed, a[i])
		default:
			return nil, ErrInvalidArgs
		}
	}

	return mixed, nil
}

func captureArg(capture interface{}, a []interface{}, f captureFlags) (nextFlags captureFlags, err error) {
	nextFlags = f

	switch p := capture.(type) {
	case *int:
		*p, err = captureInt(a[0])
	case *float64:
		*p, err = captureFloat(a[0])
	case *string:
		*p, err = captureString(a[0])
	case *time.Duration:
		*p, err = captureDuration(a[0], time.Millisecond)
	case *time.Time:
		*p, err = captureTime(a[0])
	case *[]int:
		*p, err = captureInts(a)
		nextFlags |= varargsConsumed
	case *[]float64:
		*p, err = captureFloats(a)
		nextFlags |= varargsConsumed
	case *[]string:
		*p, err = captureStrings(a)
		nextFlags |= varargsConsumed
	case *[]time.Duration:
		*p, err = captureDurations(a, time.Millisecond)
		nextFlags |= varargsConsumed
	case *[]time.Time:
		*p, err = captureTimes(a)
		nextFlags |= varargsConsumed
	case *[]interface{}:
		*p, err = captureMixed(a)
		nextFlags |= varargsConsumed
	case enum:
		switch p.p.(type) {
		case *[]string:
			*p.p.(*[]string), err = captureEnums(p.options, a)
			nextFlags |= varargsConsumed
		case *string:
			*p.p.(*string), err = captureEnum(p.options, a[0])
		}
	case duration:
		switch p.p.(type) {
		case *[]time.Duration:
			*p.p.(*[]time.Duration), err = captureDurations(a, p.unit)
			nextFlags |= varargsConsumed
		case *time.Duration:
			*p.p.(*time.Duration), err = captureDuration(a[0], p.unit)
		}
	case optional:
		if f&hasArgs != 0 {
			nextFlags, err = captureArg(p.p, a, f)
			nextFlags |= expectingOptional
		}
	default:
		err = errNotSupportedCaptureType
	}

	return
}

func Capture(a ...interface{}) error {
	captures, args, err := splitArgs(a)
	if err != nil {
		return err
	}

	if len(captures) == 0 {
		if len(args) == 0 {
			return nil
		}

		return ErrInvalidArgs
	}

	var (
		index   int
		capture interface{}
		f       captureFlags
	)

	f = hasArgs
	for index, capture = range captures {
		if len(args) == index {
			f = f &^ hasArgs
		}

		if index == len(captures)-1 {
			f |= lastPosition
		}

		if err := validateCapture(capture, f); err != nil {
			return err
		}

		if f&hasArgs != 0 {
			if f, err = captureArg(capture, args[index:], f); err != nil {
				return err
			}
		}
	}

	if f&varargsConsumed == 0 && index+1 < len(args) {
		return ErrInvalidArgs
	}

	return nil
}

func Enum(a interface{}, options ...string) interface{} {
	switch p := a.(type) {
	case *string, *[]string:
		return enum{
			options: options,
			p:       p,
		}
	case optional:
		return Optional(Enum(p.p, options...))
	default:
		return enum{invalid: true}
	}
}

func Duration(a interface{}, unit time.Duration) interface{} {
	switch p := a.(type) {
	case *time.Duration, *[]time.Duration:
		return duration{
			unit: unit,
			p:    p,
		}
	case optional:
		return Optional(Duration(p.p, unit))
	default:
		return duration{invalid: true}
	}
}

func Optional(a interface{}) interface{} {
	return optional{a}
}
