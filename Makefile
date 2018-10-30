VERSION = 0.3.0

build/blinky: BUILDFLAGS ?= -ldflags "-s -w -X main.ConfDir=/etc/blinky -X main.Version=$(VERSION)"
build/blinky: *.go
	go build -o "$@" $(BUILDFLAGS)

build/blinky-dev: BUILDFLAGS ?=
build/blinky-dev: *.go
	go build -o "$@" $(BUILDFLAGS)

build/v%.tar.gz:
	@mkdir -p build
	curl -fsL -o "$@" "https://github.com/fengb/blinky/archive/$(@F)"

build/PKGBUILD-v%: build/v%.tar.gz scripts/PKGBUILD.template
	VERSION=$(VERSION) \
	SHA256="$$(sha256sum '$<' | cut -d' ' -f1)" \
	scripts/PKGBUILD.template >"$@"

build/PKGBUILD: build/PKGBUILD-v$(VERSION)
	cp "$<" "$@"

.SECONDARY:

.PHONY: build clean pkgbuild makepkg

build: build/blinky

clean:
	rm -rf build/*

pkgbuild: build/PKGBUILD

makepkg: USER ?= nobody
makepkg: build/PKGBUILD
	chmod 777 build
	cd build && su -s /bin/bash -c "makepkg --clean" $(USER)
