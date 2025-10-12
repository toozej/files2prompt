# setup project and deps
FROM golang:1.25-bookworm AS init

WORKDIR /go/files2prompt/

COPY go.mod* go.sum* ./
RUN go mod download

COPY . ./

FROM init AS vet
RUN go vet ./...

# run tests
FROM init AS test
RUN go test -coverprofile c.out -v ./...

# build binary
FROM init AS build
ARG LDFLAGS

# Install coreutils for sleep and other utilities utilized in devcontainer
RUN apt-get update && apt-get install --no-install-recommends -y coreutils

RUN CGO_ENABLED=0 go build -ldflags="${LDFLAGS}"

# runtime image
FROM scratch
# Copy our static executable.
COPY --from=build /go/files2prompt/files2prompt /go/bin/files2prompt
# Run the binary.
ENTRYPOINT ["/go/bin/files2prompt"]
