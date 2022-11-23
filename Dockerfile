FROM golang:1.19-alpine AS handler

ENV CGO_ENABLED=0

WORKDIR /go/src/
ADD spotify-icecast /go/src/spotify-icecast
WORKDIR /go/src/spotify-icecast
RUN go mod download
RUN go build


FROM debian:buster-slim

ENV JAVA_HOME=/opt/java/openjdk
COPY --from=eclipse-temurin:11 $JAVA_HOME $JAVA_HOME
ENV PATH="${JAVA_HOME}/bin:${PATH}"

ARG LIBRESPOT_VERSION=1.6.2

RUN apt update && apt install -y pulseaudio ca-certificates wget darkice
RUN useradd -ms /bin/bash user
WORKDIR /home/user

RUN wget -O librespot.jar https://github.com/librespot-org/librespot-java/releases/download/v$LIBRESPOT_VERSION/librespot-api-$LIBRESPOT_VERSION.jar

ADD start.sh /home/user/start.sh
ADD darkice_template.cfg /home/user/darkice_template.cfg
ADD config.toml /home/user/config.toml
COPY --from=handler /go/src/spotify-icecast/spotify-icecast /home/user/spotify-icecast


RUN chmod +x /home/user/start.sh && chmod +x /home/user/spotify-icecast

USER user


ENTRYPOINT [ "./start.sh" ]

