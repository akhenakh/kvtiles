# kvtiles

kvtiles is a web server to embed and serve maps. Free map for all!

Regions are precomputed, simply pull the image and you are good to go.

Using the MVT format extracted from MBTiles, kvtiles is now using [PMTiles](https://protomaps.com/).

In short this project provides [self hosted map tiles](https://blog.nobugware.com/post/2019/self_hosted_world_maps/). 


## Docker & Kubernetes

Main goal of kvtiles is to be run in Docker/Kubernetes.

You can browse all the different regions and levels via [kvtiles docker tags](https://hub.docker.com/r/akhenakh/kvtiles/tags)

```
 docker run --rm -it -p 8080:8080 akhenakh/kvtiles:us-9-latest
```

Then point your browser to `http://yourdockerip:8080/static/`

An example deployment for kubernetes is located in `cmd/kvtilesd`.

## APIs

Tiles are available at `/tiles/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.pbf`, an optional `key` URL param can be passed to secure access to your tiles server, (use the `tilesKey` option).

Metrics are provided via Prometheus at `http://host:httpMetricsPort/metrics`.

A debug visual map is available at `http://host:httpAPIPort/static/`.

Health status is provided via gRPC `host:healthPort` or via HTTP `http://host:httpAPIPort/healthz`.

A `http://host:httpAPIPort/version` is giving you running version but also information on the dataset.


## Application usage


To serve the DB use `kvtilesd`
```
Usage of ./cmd/kvtilesd/kvtilesd:
  -allowOrigin="*": Access-Control-Allow-Origin
  -dbPath="map.pmtiles": Pmtiles path
  -healthPort=6666: grpc health port
  -httpAPIPort=8080: http API port
  -httpMetricsPort=8088: http port
  -logLevel="INFO": DEBUG|INFO|WARN|ERROR
  -tilesKey="": A key to protect your tiles access
```

## Creating PMTiles

Download the `pmtiles` binary for your system at [go-pmtiles/Releases](https://github.com/protomaps/go-pmtiles/releases).

    pmtiles convert INPUT.mbtiles OUTPUT.pmtiles
    pmtiles upload OUTPUT.mbtiles s3://my-bucket?region=us-west-2 // requires AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY env vars to be set