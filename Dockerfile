FROM debian:stable-slim

WORKDIR /

RUN apt-get update && apt-get install -y ca-certificates

COPY cloudformation-operator /usr/local/bin

CMD ["cloudformation-operator"]