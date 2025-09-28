FROM golang:1.21-alpine

WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ .

RUN go build -o pvc-audit main.go

ENTRYPOINT ["./pvc-audit"]
