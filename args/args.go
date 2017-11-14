package args

import (
	"errors"
	"log"
	"time"
)

type enum struct {
	options    []string
	valid      bool
	isVariadic bool
	p          interface{}
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
)

func splitArgs(a []interface{}) (captures []interface{}, args []interface{}, err error) {
	if len(a) == 0 {
		err = ErrInvalidArgs
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
	case *int, *float64, *string, *time.Duration:
		if f&expectingOptional != 0 {
			return errExpetingOptionalOrVararg
		}

		if f&hasArgs == 0 {
			return ErrInvalidArgs
		}
	case *[]int, *[]float64, *[]string, *[]time.Duration, *[]interface{}:
		if f&expectingSingle != 0 || f&lastPosition == 0 {
			return errNotVariadicPosition
		}
	case enum:
		if !p.valid {
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

func captureDuration(a interface{}) (time.Duration, error) {
	switch v := a.(type) {
	case int:
		return time.Duration(v) * time.Millisecond, nil
	case float64:
		scale := float64(time.Millisecond) / float64(time.Nanosecond)
		return time.Duration(v*scale) * time.Millisecond / time.Duration(scale), nil
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

func captureDurations(a []interface{}) ([]time.Duration, error) {
	var durations []time.Duration
	for i := range a {
		log.Println(a[i])
		v, err := captureDuration(a[i])
		if err != nil {
			return nil, ErrInvalidArgs
		}

		durations = append(durations, v)
	}

	return durations, nil
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
		*p, err = captureDuration(a[0])
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
		*p, err = captureDurations(a)
		nextFlags |= varargsConsumed
	case *[]interface{}:
		*p, err = captureMixed(a)
		nextFlags |= varargsConsumed
	case enum:
		if p.isVariadic {
			*p.p.(*[]string), err = captureEnums(p.options, a)
			nextFlags |= varargsConsumed
		} else {
			*p.p.(*string), err = captureEnum(p.options, a[0])
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
	)

	f := hasArgs
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
	case *string:
		return enum{
			options: options,
			valid:   true,
			p:       p,
		}
	case *[]string:
		return enum{
			options:    options,
			valid:      true,
			isVariadic: true,
			p:          p,
		}
	case optional:
		return Optional(Enum(p.p, options...))
	default:
		return enum{}
	}
}

func Optional(a interface{}) interface{} {
	return optional{a}
}

func Duration(a interface{}, unit time.Duration) interface{} {
	return nil
}
