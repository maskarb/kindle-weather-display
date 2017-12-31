#!/bin/sh -ex

/opt/weather-script
rsvg-convert --background-color=white -o weather-script-output.png output.svg
pngcrush -c 0 -ow weather-script-output.png
cp -f weather-script-output.png /var/lib/www/weather-script-output.png
