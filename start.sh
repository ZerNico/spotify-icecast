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

echo "Starting spotify-icecast and librespot..."
./spotify-icecast & ./librespot -q --name $SPOTIFY_DEVICE_NAME -u $SPOTIFY_USERNAME -p $SPOTIFY_PASSWORD --backend pulseaudio --bitrate 320 --onevent ./event.sh --cache /tmp/librespot