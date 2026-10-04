FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /ssh-portfolio ./cmd/ssh-portfolio

FROM alpine:3.22
COPY --from=build /ssh-portfolio /usr/local/bin/ssh-portfolio
RUN mkdir -p /data
# The host key lives on a volume so it survives deploys (otherwise every
# visitor gets a scary "REMOTE HOST IDENTIFICATION HAS CHANGED" warning).
EXPOSE 22 8080
ENTRYPOINT ["ssh-portfolio", "-port", "22", "-key", "/data/host_ed25519", "-http", ":8080"]
