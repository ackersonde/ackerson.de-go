#!/bin/bash

ssh core@ackerson.de CIRCLE_BUILD_NUM=$CIRCLE_BUILD_NUM <<-'ENDSSH'
	docker pull blauerdrachen/ackerson.de:vc$CIRCLE_BUILD_NUM
	if [ $? -eq 0 ]
	then
		docker stop ackerson.de && docker rm ackerson.de
		docker run -d -p 80:3001 -v /opt:/opt -e ackSecret=$ackSecret -e ackPoems=$ackPoems -e ackWunder=$ackWunder -e ackMongo=$ackMongo --name ackerson.de blauerdrachen/ackerson.de:vc$CIRCLE_BUILD_NUM
	fi
ENDSSH
