package main

// linker constants
var (
	ConfDir string
	Version string
)

func main() {
	if ConfDir == "" {
		ConfDir = "etc"
	}

	conf, err := NewConf(ConfDir)
	if err != nil {
		panic(err)
	}

	go Refresher(conf)
	Serve(conf)
}
