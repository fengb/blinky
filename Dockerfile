ARG IMAGE=base/archlinux
FROM ${IMAGE}:2018.10.01

RUN pacman --noconfirm \
           -Sy fakeroot gcc make git go inotify-tools \
    && rm -rf \
         /var/cache/pacman/pkg/* \
         /var/lib/pacman/sync/* \
         /etc/pacman.d/mirrorlist.pacnew

ENV GOPATH /opt/go
RUN mkdir -p $GOPATH/src/blinky
WORKDIR $GOPATH/src/blinky

CMD bash
