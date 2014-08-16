package valve

import (
	"fmt"
)

func tryAndCatch(fn func() error) error {
	var outErr error
	(func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				outErr = err
			}
		}()

		outErr = fn()
	})()
	return outErr
}

func tryNoCatch(fn func() error) error {
	return fn()
}

// This can be changed to enable panics.
var Try = tryNoCatch
