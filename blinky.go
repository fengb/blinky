package main

func main() {
	conf, err := LoadConfDir("etc")
	if err != nil {
		panic(err)
	}

	pac, err := NewPac("/etc/pacman.conf")
	if err != nil {
		panic(err)
	}

	Serve(conf, pac)
}
