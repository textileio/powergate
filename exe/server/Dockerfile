FROM golang:1.13.7 as builder
RUN apt-get update && apt-get install -y mesa-opencl-icd ocl-icd-opencl-dev
RUN mkdir /app 
WORKDIR /app 
COPY extern extern
COPY go.mod go.sum ./
RUN go mod download
COPY . . 
RUN GOOS=linux go build -o main exe/server/main.go && \
go run github.com/GeertJohan/go.rice/rice append --exec main -i ./build

FROM ubuntu
RUN apt-get update && apt-get install -y mesa-opencl-icd ocl-icd-opencl-dev
COPY --from=builder /app/api/server/ip2location-ip4.bin /app/ip2location-ip4.bin
COPY --from=builder /app/main /app/main
WORKDIR /app 
ENTRYPOINT ["./main"]