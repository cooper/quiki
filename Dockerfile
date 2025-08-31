FROM golang:1.23-alpine AS builder
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o quiki .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/quiki /bin/quiki
EXPOSE 8080
USER 65534:65534
ENTRYPOINT ["/bin/quiki", "-w"]