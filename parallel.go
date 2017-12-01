package main

import "github.com/hashicorp/go-multierror"

func Parallel(fns ...func() error) (result error) {
	cerrs := make(chan error, len(fns))

	for _, fn := range fns {
		go func(fn func() error) {
			cerrs <- fn()
		}(fn)
	}

	for _, _ = range fns {
		err := <-cerrs
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}
