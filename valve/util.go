// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
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
var Try = tryAndCatch
