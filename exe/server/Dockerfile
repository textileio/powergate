FROM golang:1.13.7 as builder
RUN apt-get update && apt-get install -y mesa-opencl-icd ocl-icd-opencl-dev
RUN mkdir /app 
WORKDIR /app 
COPY extern extern
COPY go.mod go.sum ./
RUN go mod download
COPY . . 
RUN GOOS=linux go build -o fil-toolsd exe/server/main.go && \
go run github.com/GeertJohan/go.rice/rice append --exec fil-toolsd -i ./build

FROM ubuntu
RUN apt-get update && apt-get install -y mesa-opencl-icd ocl-icd-opencl-dev
COPY --from=builder /app/api/server/ip2location-ip4.bin /app/ip2location-ip4.bin
COPY --from=builder /app/fil-toolsd /app/fil-toolsd
WORKDIR /app 
ENTRYPOINT ["./fil-toolsd"]