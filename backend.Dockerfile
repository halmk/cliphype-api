FROM golang:alpine3.16

RUN apk update

# Install packages
RUN apk add python3 make g++ go
RUN apk add tmux vim git

