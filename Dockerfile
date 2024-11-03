FROM golang:latest as build

WORKDIR /build
COPY . . 
RUN go build -o /app main.go

FROM alpine:latest as run
WORKDIR /app
COPY --from=build /app /app/app
RUN apk --update add libc6-compat
CMD [ "/app/app", "start" ]