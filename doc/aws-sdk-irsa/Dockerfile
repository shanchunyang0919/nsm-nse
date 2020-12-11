# first stage - prepare a binary file
FROM golang:1.9 AS builder

RUN go version

WORKDIR /src

RUN set -x && \ 
 go get github.com/aws/aws-sdk-go/aws && \
 go get github.com/aws/aws-sdk-go/aws/session && \
 go get github.com/aws/aws-sdk-go/service/s3

COPY ./aws-sdk/main.go /src
RUN CGO_ENABLED=0 go build -o /demo

CMD ["/demo"]

# second stage - takes only the binary
#FROM busybox

#COPY --from=builder /demo .

#CMD ["/demo"]
