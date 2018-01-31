FROM golang:1.9 as builder
COPY . /
WORKDIR /
RUN go get k8s.io/client-go/...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine
COPY --from=builder /main /main
CMD ["/main"]
