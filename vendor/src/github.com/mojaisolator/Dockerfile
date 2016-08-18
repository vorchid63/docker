# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.5

# get the godep tool to take care of dependencies
RUN go get github.com/tools/godep

# Build the isolhttps command inside the container.
COPY . /go/src/github.com/mojaisolator/plugin

WORKDIR /go/src/github.com/mojaisolator/plugin
RUN godep go install -v

# Run the outyet command by default when the container starts.
ENTRYPOINT ["mojaisolator"]

# Document that the service listens on port 8080.
EXPOSE 8080