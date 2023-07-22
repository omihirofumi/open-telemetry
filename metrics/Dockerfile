FROM golang:1.19 as build

COPY . /go/src/metrics

WORKDIR /go/src/metrics

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o metrics

FROM scratch as image

COPY --from=build /go/src/metrics/metrics .

EXPOSE 2222

CMD ["/metrics"]