PREFIX ?=
VERSION = 0.2.0

build/blinky: BUILDFLAGS ?= -ldflags "-s -w -X main.ConfDir=/etc/blinky -X main.Version=$(VERSION)"
build/blinky: *.go
	go build -o "$@" $(BUILDFLAGS)

build/blinky-dev: BUILDFLAGS ?=
build/blinky-dev: *.go
	go build -o "$@" $(BUILDFLAGS)

build/v%.tar.gz:
	@mkdir -p build
	curl -fsL -o "$@" "https://github.com/fengb/blinky/archive/$(@F)"

build/PKGBUILD-v%: build/v%.tar.gz scripts/PKGBUILD
	scripts/expand_vars VERSION=$(VERSION) SHA256=$$(scripts/sha256 "$<") <scripts/PKGBUILD >"$@"

build/PKGBUILD: build/PKGBUILD-v$(VERSION)
	cp "$<" "$@"

.SECONDARY:

.PHONY: build clean pkgbuild makepkg install uninstall

build: build/blinky

clean:
	rm -rf build/*

pkgbuild: build/PKGBUILD

makepkg: USER ?= nobody
makepkg: build/PKGBUILD
	chmod 777 build
	cd build && su -s /bin/bash -c "makepkg --clean" $(USER)

install: build
	install -D -m0644 -t$(PREFIX)/etc/blinky etc/*
	install -D -m0755 -t$(PREFIX)/usr/bin build/blinky
	install -D -m0644 -t$(PREFIX)/usr/lib/systemd/system systemd/*

uninstall:
	rm -r $(PREFIX)/etc/blinky
	rm $(PREFIX)/usr/bin/blinky
	rm $(PREFIX)/usr/lib/systemd/system/blinky*.service
