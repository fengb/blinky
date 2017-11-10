PREFIX ?= /
VERSION = 0.0.1

build/blinky:
	go build -o build/blinky

build/v%.tar.gz:
	@mkdir -p build
	curl -fsL -o $@ https://github.com/fengb/blinky/archive/v$*.tar.gz

build/PKGBUILD: build/v$(VERSION).tar.gz
	VERSION=$(VERSION) TARBALL=build/v$(VERSION).tar.gz \
	scripts/pkgbuild build/PKGBUILD

.PHONY: clean install uninstall package

clean:
	rm -r build

package: build/PKGBUILD

install: build/blinky
	install -D -m0755 build/blinky $(PREFIX)/usr/bin/blinky

uninstall:
	rm $(PREFIX)/usr/bin/blinky
