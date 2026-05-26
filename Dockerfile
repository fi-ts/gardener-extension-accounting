FROM golang:1.26 AS builder

WORKDIR /go/src/github.com/fi-ts/gardener-extension-accounting
COPY . .
RUN make install \
 && strip /go/bin/gardener-extension-accounting

FROM alpine:3.22
WORKDIR /
COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-accounting /gardener-extension-accounting
CMD ["/gardener-extension-accounting"]
