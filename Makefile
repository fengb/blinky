PREFIX ?= /

.DEFAULT = build

build/blinky:
	go build -o build/blinky

.PHONY: build install uninstall

build: blinky

install:
	install -D -m0755 blinky $(PREFIX)/usr/bin/blinky

uninstall:
	rm $(PREFIX)/usr/bin/blinky
