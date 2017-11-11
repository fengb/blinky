ARG IMAGE=base/archlinux
FROM ${IMAGE}:latest

RUN pacman --noconfirm \
           -Syu gcc make git go \
    && rm -rf \
         /var/cache/pacman/pkg/* \
         /var/lib/pacman/sync/* \
         /etc/pacman.d/mirrorlist.pacnew

ENV GOPATH /opt/go
RUN mkdir -p $GOPATH/src/blinky
WORKDIR $GOPATH/src/blinky

CMD bash
