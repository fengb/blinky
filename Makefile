PREFIX ?=
export VERSION = 0.1.1
LDFLAGS = "-s -w -X main.ConfDir=/etc/blinky -X main.Version=$(VERSION)"

build/blinky: *.go
	go build -o "$@" -ldflags $(LDFLAGS)

build/v%.tar.gz:
	@mkdir -p build
	curl -fsL -o "$@" "https://github.com/fengb/blinky/archive/$(@F)"

build/PKGBUILD-v%: build/v%.tar.gz
	scripts/pkgbuild "$<" "$@"

build/PKGBUILD: build/PKGBUILD-v$(VERSION)
	cp "$<" "$@"

.SECONDARY:

.PHONY: build clean pkgbuild install uninstall

build: build/blinky

clean:
	rm -rf build/*

pkgbuild: build/PKGBUILD

install: build
	install -D -m0644 -t$(PREFIX)/etc/blinky etc/*
	install -D -m0755 -t$(PREFIX)/usr/bin build/blinky
	install -D -m0644 -t$(PREFIX)/usr/lib/systemd/system systemd/*

uninstall:
	rm -r $(PREFIX)/etc/blinky
	rm $(PREFIX)/usr/bin/blinky
	rm $(PREFIX)/usr/lib/systemd/system/blinky*.service
