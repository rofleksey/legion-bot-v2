FROM golang:1.24-alpine AS builder
WORKDIR /opt
COPY . /opt/
RUN go mod download
RUN go build -o ./legion-bot-v2

FROM alpine
ARG ENVIRONMENT=production
WORKDIR /opt
RUN apk update && apk add --no-cache curl ca-certificates
COPY --from=builder /opt/legion-bot-v2 /opt/legion-bot-v2
COPY ./static /opt/static
EXPOSE 8080
CMD [ "./legion-bot-v2" ]
