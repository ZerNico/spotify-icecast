FROM rustlang/rust:nightly AS librespot

RUN apt update && apt install -y pulseaudio libasound2-dev libavahi-compat-libdnssd-dev pkg-config

WORKDIR /usr/src/
RUN git clone https://github.com/librespot-org/librespot
WORKDIR /usr/src/librespot
RUN cargo build --release --features pulseaudio-backend

FROM ubuntu:latest

RUN apt update && apt install -y pulseaudio ca-certificates curl darkice
RUN useradd -ms /bin/bash user

ADD start.sh /home/user/start.sh
ADD darkice_template.cfg /home/user/darkice_template.cfg
ADD event.sh /home/user/event.sh
COPY --from=librespot /usr/src/librespot/target/release/librespot /home/user/librespot


RUN chmod +x /home/user/start.sh \
    && chmod +x /home/user/librespot \
    && chmod +x /home/user/event.sh

USER user
WORKDIR /home/user

ENTRYPOINT [ "./start.sh" ]

