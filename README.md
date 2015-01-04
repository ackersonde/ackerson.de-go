Using taking the Dockerfile above:
1. docker build -t="blauerdrachen/ackerson.de-go" --no-cache .
2. docker run -d -p 80:3000 blauerdrachen/ackerson.de-go

