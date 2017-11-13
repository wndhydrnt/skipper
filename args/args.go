package args

import "errors"

var (
	ErrInvalidArgs = errors.New("invalid args")

	errArgPositionType     = errors.New("last argument must be of type []interface{} or nil")
	errNotSupportedArgType = errors.New("not supported capture type")
	errVariadicNotLast     = errors.New("variadic must be the last arg")
	errInvalidEnum         = errors.New("invalid enum")
)

type enum struct {
	options    []string
	valid      bool
	isVariadic bool
	p          interface{}
}

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

func validateCapture(capture interface{}, lastCapture, hasArg bool) error {
	switch p := capture.(type) {
	case *int, *float64, *string:
		if !hasArg {
			return ErrInvalidArgs
		}
	case *[]int, *[]float64, *[]string, *[]interface{}:
		if !lastCapture {
			return errVariadicNotLast
		}
	case enum:
		if !p.valid {
			return errInvalidEnum
		}

		if p.isVariadic && !lastCapture {
			return ErrInvalidArgs
		}

		if !p.isVariadic && !hasArg {
			return ErrInvalidArgs
		}
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
		index           int
		capture         interface{}
		varargsConsumed bool
	)

	for index, capture = range captures {
		isLast := index == len(captures)-1
		hasArg := len(args) > index
		if err := validateCapture(capture, isLast, hasArg); err != nil {
			return err
		}

		var err error
		switch p := capture.(type) {
		case *int:
			*p, err = captureInt(args[index])
		case *float64:
			*p, err = captureFloat(args[index])
		case *string:
			*p, err = captureString(args[index])
		case *[]int:
			*p, err = captureInts(args[index:])
			varargsConsumed = true
		case *[]float64:
			*p, err = captureFloats(args[index:])
			varargsConsumed = true
		case *[]string:
			*p, err = captureStrings(args[index:])
			varargsConsumed = true
		case *[]interface{}:
			*p, err = captureMixed(args[index:])
			varargsConsumed = true
		case enum:
			if p.isVariadic {
				*p.p.(*[]string), err = captureEnums(p.options, args[index:])
				varargsConsumed = true
			} else {
				*p.p.(*string), err = captureEnum(p.options, args[index])
			}
		default:
			err = errNotSupportedArgType
		}

		if err != nil {
			return err
		}
	}

	if !varargsConsumed && index+1 < len(args) {
		return ErrInvalidArgs
	}

	return nil
}

func Enum(a interface{}, options ...string) interface{} {
	e := enum{options: options}
	switch p := a.(type) {
	case *string:
		e.valid = true
		e.p = p
	case *[]string:
		e.valid = true
		e.isVariadic = true
		e.p = p
	}

	return e
}
