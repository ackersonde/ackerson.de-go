#!/bin/bash

ssh core@ackerson.de <<-'ENDSSH'
	docker pull blauerdrachen/ackerson.de:vc$CIRCLE_BUILD_NUM
	docker stop ackerson.de && docker rm ackerson.de
	docker run -d -p 80:3001 -v /opt:/opt -e ackSecret=$ackSecret -e ackPoems=$ackPoems -e ackWunder=$ackWunder -e ackMongo=$ackMongo --name ackerson.de blauerdrachen/ackerson.de:vc$CIRCLE_BUILD_NUM
ENDSSH
