FROM golang:1.13
MAINTAINER Stanley <grimmh6919@hotmail.com.tw>

WORKDIR $GOPATH/src

RUN apt-get update && apt-get -y install nodejs npm git vim build-essential unzip curl wget && \
    npm install pm2 -g
RUN go get -u github.com/derekparker/delve/cmd/dlv

# clear apt-get cache
RUN apt-get autoclean && apt-get -y autoremove && rm -rf /var/lib/apt/lists/*

EXPOSE 8000 9001