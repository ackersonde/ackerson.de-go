[![Circle CI](https://circleci.com/gh/danackerson/ackerson.de-go.svg?style=shield&circle-token=3ad6694a5592b15aef77eeb7051a7b6c61d1c56f)](https://circleci.com/gh/danackerson/ackerson.de-go)

# Installation and Development
0. go get github.com/pilu/fresh
0. go get github.com/urfave/negroni
0. go get github.com/goincremental/negroni-session
0. go get gopkg.in/mgo.v2
0. cd ~/dev/danackerson/ackerson.de-go/
0. vi /opt/creds.txt (with appropriate values!)
0. `fresh` (launch app refresher for dev)
0. http://localhost:8080 (now code and fresh builds in the background)

# Building
0. cd ~/dev/danackerson/ackerson.de-go/
0. docker build -t="blauerdrachen/ackerson.de" --no-cache .
0. docker login
0. docker push blauerdrachen/ackerson.de
0. docker run -d -p 80:8080 -p 443:8443 --name="ackerson.de" blauerdrachen/ackerson.de
0. curl http://ackerson.de/

# Running
Automatic startup on CoreOS:
```
$ docker stop ackerson.de

$ sudo vi -r /etc/systemd/system/ackerson-de.service
[Unit]
Description=AckersonHomepage
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
ExecStart=/usr/bin/docker run ackerson.de

[Install]
WantedBy=multi-user.target

$ sudo systemctl enable /etc/systemd/system/ackerson-de.service
$ sudo systemctl start ackerson-de.service
$ sudo reboot
```

Final check:

`curl http://ackerson.de`
