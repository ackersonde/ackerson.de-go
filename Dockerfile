FROM iron/base
RUN mkdir -p /root/certs
WORKDIR /app
COPY server /app/
ENTRYPOINT ["./server"]
