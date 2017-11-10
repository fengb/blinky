package main

func main() {
	conf, err := DefaultConf()
	if err != nil {
		panic(err)
	}

	pac, err := NewPac("/etc/pacman.conf")
	if err != nil {
		panic(err)
	}

	watch, err := pac.Watch()
	if err != nil {
		panic(err)
	}

	Serve(conf, watch)
}
