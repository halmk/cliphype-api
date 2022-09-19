FROM node:alpine

RUN apk update

# Install packages
RUN apk add python3 make g++ go
RUN apk add tmux vim git

