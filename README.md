# Installation and Development
0. go get github.com/codegangsta/gin
0. go get github.com/codegangsta/negroni
0. go get github.com/goincremental/negroni-session
0. go get gopkg.in/mgo.v2
0. cd ~/dev/danackerson/ackerson.de-go/
0. vi /opt/creds.txt (with appropriate values!)
0. `gin` (launch app refresher for dev)
0. http://localhost:3001 (now code and gin builds in the background)

# Building
0. cd ~/dev/danackerson/ackerson.de-go/
0. docker build -t="blauerdrachen/ackerson.de" --no-cache .
0. docker login
0. docker push blauerdrachen/ackerson.de
0. docker run -d -p 80:3001 -v /opt:/opt --name="ackerson.de" blauerdrachen/ackerson.de
0. curl http://ackerson.de/

# Running
Automatic startup on CoreOS:
```
$ docker stop ackerson.de
$ sudo vi /opt/creds.txt
mongo=mongodb://XYZ.mongolab.com:123/abc
secret=[secret]
wunderground=[api]
poem=[param]

$ sudo vi -r /etc/systemd/system/ackerson-de.service
[Unit]
Description=AckersonHomepage
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
ExecStart=/usr/bin/docker start ackerson.de

[Install]
WantedBy=multi-user.target

$ sudo systemctl enable /etc/systemd/system/ackerson-de.service
$ sudo systemctl start ackerson-de.service
$ sudo reboot
```

Final check:

`curl http://ackerson.de`
