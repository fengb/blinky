PREFIX ?= /
VERSION = 0.0.1

build/blinky:
	go build -o build/blinky

build/blinky.tar.gz: *.go
	@mkdir -p build
	git ls-files | xargs tar --transform 's:^:blinky-$(VERSION)/:' -czf build/blinky.tar.gz

build/PKGBUILD: build/blinky.tar.gz
	VERSION=$(VERSION) TARBALL=build/blinky.tar.gz \
	scripts/pkgbuild build/PKGBUILD

.PHONY: clean install uninstall package

clean:
	rm -r build

package: build/blinky.tar.gz build/PKGBUILD

install:
	install -D -m0755 blinky $(PREFIX)/usr/bin/blinky

uninstall:
	rm $(PREFIX)/usr/bin/blinky
