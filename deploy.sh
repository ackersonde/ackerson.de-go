#!/bin/bash

ssh core@ackerson.de <<-'ENDSSH'
	touch /tmp/this
	ls -lrt /tmp
ENDSSH
