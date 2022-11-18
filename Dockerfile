FROM rustlang/rust:nightly AS librespot-token

RUN apt update && apt install -y pulseaudio libasound2-dev libavahi-compat-libdnssd-dev pkg-config

WORKDIR /usr/src/
ADD librespot-token /usr/src/librespot-token
WORKDIR /usr/src/librespot-token
RUN cargo build --release



FROM rustlang/rust:nightly AS librespot

RUN apt update && apt install -y pulseaudio libasound2-dev libavahi-compat-libdnssd-dev pkg-config

WORKDIR /usr/src/
RUN git clone https://github.com/librespot-org/librespot
WORKDIR /usr/src/librespot
RUN cargo build --release --no-default-features --features pulseaudio-backend



FROM golang:1.19-alpine AS handler

ENV CGO_ENABLED=0

WORKDIR /go/src/
ADD spotify-icecast /go/src/spotify-icecast
WORKDIR /go/src/spotify-icecast
RUN go mod download
RUN go build



FROM debian:stable-slim

RUN apt update \
    && apt install -y pulseaudio alsa-utils darkice ca-certificates curl
RUN useradd -ms /bin/bash user

ADD start.sh /home/user/start.sh
ADD darkice_template.cfg /home/user/darkice_template.cfg
ADD event.sh /home/user/event.sh
COPY --from=librespot-token /usr/src/librespot-token/target/release/librespot-token /home/user/librespot-token
COPY --from=librespot /usr/src/librespot/target/release/librespot /home/user/librespot
COPY --from=handler /go/src/spotify-icecast/spotify-icecast /home/user/spotify-icecast


RUN chmod +x /home/user/start.sh \
    && chmod +x /home/user/librespot

USER user
WORKDIR /home/user

ENTRYPOINT [ "./start.sh" ]

