FROM golang:alpine AS build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o auth

FROM alpine:3.19
COPY --from=build_container /app/auth /usr/bin
RUN mkdir security
COPY ./security/blacklist.txt ./security/
EXPOSE 8081
ENTRYPOINT ["auth"]
