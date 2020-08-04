FROM golang:1.14.2-buster as builder

RUN mkdir /app 
WORKDIR /app 

COPY go.mod go.sum ./
RUN go mod download
# this can go away after i figure out how to keep the dep in go.mod
RUN go get github.com/ahmetb/govvv
COPY . .
RUN pkg="$(go list ./buildinfo)" && govvvflags="$(govvv -flags -pkg $pkg)" && CGO_ENABLED=0 GOOS=linux go build -ldflags="$govvvflags" -o powd cmd/powd/main.go

FROM alpine
COPY --from=builder /app/iplocation/maxmind/GeoLite2-City.mmdb /app/GeoLite2-City.mmdb
COPY --from=builder /app/powd /app/powd
WORKDIR /app 
ENTRYPOINT ["./powd"]
