Using the Dockerfile above:
1. docker build -t="blauerdrachen/ackerson.de" --no-cache .
2. docker push blauerdrachen/ackerson.de
3. docker run -d -p 80:3000 --name="ackerson.de" blauerdrachen/ackerson.de
