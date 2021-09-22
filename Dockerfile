FROM quay.io/prometheus/golang-builder:latest as build

ENV CGO_ENABLED 0
ADD . /app/node_exporter_hostname
WORKDIR /app/node_exporter_hostname

RUN go get
RUN go build -trimpath -ldflags="-s -w"
RUN chmod +x node_exporter_hostname

FROM prom/node-exporter:latest

COPY --from=build /app/node_exporter_hostname/node_exporter_hostname /bin/node_exporter_hostname

ENTRYPOINT [ "/bin/node_exporter_hostname" ]
