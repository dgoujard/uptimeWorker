############################
# STEP 1 build executable binary
############################
FROM golang:1.14.6-alpine3.12 as builder
# Create appuser.
ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN apk add --update --no-cache ca-certificates git tzdata && update-ca-certificates
# We create an /app directory within our
# image that will hold our application source
# files
RUN mkdir /app
# We copy everything in the root directory
# into our /app directory
ADD . /app

WORKDIR /app
# We specify that we now wish to execute
# any further commands inside our /app
# directory

RUN go install github.com/markbates/pkger/cmd/pkger
RUN go get -d -v  github.com/dgoujard/uptimeWorker/cmd/uptimeworker

# Using go mod.
# RUN go mod download
# RUN go mod verify

# Build the binary.
RUN pkger -o cmd/uptimeworker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -ldflags="-w -s" -o uptimeworker  github.com/dgoujard/uptimeWorker/cmd/uptimeworker && chmod +x uptimeworker

############################
# STEP 3 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/dnsapi /app/dnsapi

# Use an unprivileged user.
USER appuser:appuser

EXPOSE 8080

ENTRYPOINT ["/app/uptimeworker"]
