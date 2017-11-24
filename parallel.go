package main

func Parallel(fns ...func() error) error {
	cerrs := make(chan error, len(fns))

	for _, fn := range fns {
		go func(fn func() error) {
			cerrs <- fn()
		}(fn)
	}

	for _, _ = range fns {
		err := <-cerrs
		if err != nil {
			return err
		}
	}

	return nil
}
