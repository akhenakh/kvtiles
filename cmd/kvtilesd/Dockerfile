FROM gcr.io/distroless/static

WORKDIR /root/
COPY map.db .
ADD static /root/static
COPY kvtilesd .
COPY grpc_health_probe .
EXPOSE 6666 8088 8080
ENTRYPOINT ["./kvtilesd"]
