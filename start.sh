#!/bin/sh
echo "Starting pulseaudio..."
rm -rf .config/pulse
rm -rf /tmp/pulse*
pulseaudio -D --exit-idle-time=-1

sleep 2

while read line
do
    eval echo "$line"
done < "./darkice_template.cfg" > "./darkice.cfg"

sed -ri "s/SPOTIFY_USERNAME/$SPOTIFY_USERNAME/" ./config.toml
sed -ri "s/SPOTIFY_PASSWORD/$SPOTIFY_PASSWORD/" ./config.toml
sed -ri "s/SPOTIFY_DEVICE_NAME/$SPOTIFY_DEVICE_NAME/" ./config.toml

echo "Starting spotify-icecast and librespot..."
java -jar librespot.jar & sleep 10 && ./spotify-icecast