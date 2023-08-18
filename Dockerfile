# syntax=docker/dockerfile:1

##
## STEP 1 - BUILD
##

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.21-alpine AS build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image
COPY go.mod go.sum ./

# download Go modules and dependencies
RUN go mod download

# copy directory files i.e all files ending with .go
COPY *.go ./

# compile application
RUN go build -o /gluebit

##
## STEP 2 - DEPLOY
##
FROM scratch

WORKDIR /

COPY --from=build /gluebit /gluebit

ENTRYPOINT ["/gluebit"]