Using the Dockerfile above:

0. docker build -t="blauerdrachen/ackerson.de" --no-cache .
0. docker login
0. docker push blauerdrachen/ackerson.de
0. docker run -d -p 80:3000 --name="ackerson.de" blauerdrachen/ackerson.de
0. curl http://ackerson.de/

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
ExecStart=/usr/bin/docker start ackerson.de

[Install]
WantedBy=multi-user.target

$ sudo systemctl enable /etc/systemd/system/ackerson-de.service
$ sudo systemctl start ackerson-de.service
$ sudo reboot
```

Final check:

`curl http://ackerson.de`
