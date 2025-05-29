####################  BUILD STAGE  ####################
FROM golang:1.23-alpine AS build
WORKDIR /src

# Static builds for *everything* we compile
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

# gitleaks – static, no libc needed
RUN go install github.com/zricethezav/gitleaks/v8@latest

# keysweep-scanner
COPY scanner-cli/ ./scanner-cli
RUN go build -o /bin/keysweep-scanner ./scanner-cli


###################  RUNTIME STAGE  ###################
FROM alpine:3.19
RUN apk add --no-cache git

# Make sure the rules directory exists
RUN mkdir -p /workspace

COPY --from=build /go/bin/gitleaks       /bin/
COPY --from=build /bin/keysweep-scanner  /bin/
COPY scanner-cli/gitleaks.toml           /workspace/gitleaks.toml  
COPY action/entrypoint.sh                /entrypoint.sh
