FROM golang:1.24-alpine AS builder
WORKDIR /opt
COPY . /opt/
RUN go mod download
RUN go build -o ./legion-bot-v2

FROM node:18 AS frontend
WORKDIR /opt
COPY ./frontend/ /opt/
RUN npm i && npm run build

FROM alpine
ARG ENVIRONMENT=production
WORKDIR /opt
RUN apk update && apk add --no-cache curl ca-certificates
COPY --from=builder /opt/legion-bot-v2 /opt/legion-bot-v2
COPY --from=frontend /opt/dist /opt/frontend/dist
EXPOSE 8080
CMD [ "./legion-bot-v2" ]
