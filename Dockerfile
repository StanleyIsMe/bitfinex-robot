FROM golang:1.13
MAINTAINER Stanley <grimmh6919@hotmail.com.tw>
ENV TZ=Asia/Taipei
WORKDIR $GOPATH/src/robot

RUN apt-get update && apt-get -y install nodejs npm git vim build-essential unzip curl wget && \
    npm install pm2 -g

COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o robot

# clear apt-get cache
RUN apt-get autoclean && apt-get -y autoremove && rm -rf /var/lib/apt/lists/*
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

#EXPOSE 8000 8080 9001
CMD [ "pm2", "start", "app.json", "--no-daemon"]