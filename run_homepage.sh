#!/bin/sh
apk --no-cache add curl mailcap tzdata
cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime
echo "Europe/Berlin" > /etc/timezone
./tmp/homepage
