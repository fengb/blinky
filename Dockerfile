ARG IMAGE
FROM $IMAGE

RUN pacman --noconfirm \
           -Sy fakeroot gcc make git go inotify-tools

ENV GOPATH /opt/go
RUN mkdir -p $GOPATH/src/blinky
WORKDIR $GOPATH/src/blinky

CMD bash
