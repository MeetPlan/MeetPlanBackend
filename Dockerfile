FROM golang:1.22-alpine AS builder

COPY . /app

WORKDIR /app

# Add gcc
RUN apk add build-base

RUN go mod download && \
    go env -w GOFLAGS=-mod=mod && \
    go get . && \
    go build -v .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/MeetPlanBackend ./MeetPlanBackend

# Copy fonts, school icons and offical documents
COPY ./fonts /app/fonts
COPY ./icons /app/icons
COPY ./officialdocs /app/officialdocs

EXPOSE 80
CMD [ "./MeetPlanBackend", "--useenv" ]