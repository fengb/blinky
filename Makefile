PREFIX ?=
VERSION = 0.0.1
LDFLAGS = "-s -w -X main.ConfDir=/etc/blinky -X main.Version=$(VERSION)"

build/blinky: *.go
	go build -o build/blinky -ldflags $(LDFLAGS)

build/v%.tar.gz:
	@mkdir -p build
	curl -fsL -o $@ https://github.com/fengb/blinky/archive/v$*.tar.gz

build/PKGBUILD: build/v$(VERSION).tar.gz
	VERSION=$(VERSION) TARBALL=build/v$(VERSION).tar.gz \
	scripts/pkgbuild build/PKGBUILD

.PHONY: clean install uninstall package

clean:
	rm -rf build/*

package: build/PKGBUILD

install: build/blinky
	install -D -m0644 -t$(PREFIX)/etc/blinky etc/*
	install -D -m0755 -t$(PREFIX)/usr/bin build/blinky
	install -D -m0644 -t$(PREFIX)/usr/lib/systemd/system systemd/*

uninstall:
	rm -r $(PREFIX)/etc/blinky
	rm $(PREFIX)/usr/bin/blinky
	rm $(PREFIX)/usr/lib/systemd/system/blinky*.service
