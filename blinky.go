package main

func main() {
	pac, err := NewPac("/etc/pacman.conf")
	if err != nil {
		panic(err)
	}

	watch, err := pac.Watch()
	if err != nil {
		panic(err)
	}

	Serve(watch)
}
