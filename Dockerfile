FROM alpine:latest
RUN apk --no-cache add curl mailcap tzdata
RUN cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime
RUN echo "Europe/Berlin" > /etc/timezone
ADD homepage /app/
ADD last_docker_push /app/
WORKDIR /app
ENTRYPOINT ["/app/homepage"]
