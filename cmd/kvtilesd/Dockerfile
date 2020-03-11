FROM gcr.io/distroless/base

WORKDIR /root/
COPY kvtilesd .
COPY grpc_health_probe .
COPY map.db .
ADD static /root/static
EXPOSE 6666 8088 8080
ENTRYPOINT ["./kvtilesd"]
