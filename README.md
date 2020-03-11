# kvtiles

kvtiles is a web server to embed and serve maps.

Using the MVT format extracted from MBTiles, kvtiles is using a key value storage to speed up queries.

In short this project provides [self hosted map tiles](https://blog.nobugware.com/post/2019/self_hosted_world_maps/). 

## Usage

To transform an MBTiles into an embedded DB use `mbtilestokv`
```
Usage of ./cmd/mbtilestokv/mbtilestokv:
  -centerLat=48.8: Latitude center used for the debug map
  -centerLng=2.2: Longitude center used for the debug map
  -dbPath="./map.db": db path out
  -logLevel="INFO": DEBUG|INFO|WARN|ERROR
  -maxZoom=9: max zoom used for the debug map
  -tilesPath="./hawaii.mbtiles": mbtiles file path
```

To serve the DB use `kvtilesd`
```
Usage of ./cmd/kvtilesd/kvtilesd:
  -allowOrigin="*": Access-Control-Allow-Origin
  -dbPath="map.db": Database path
  -healthPort=6666: grpc health port
  -httpAPIPort=8080: http API port
  -httpMetricsPort=8088: http port
  -logLevel="INFO": DEBUG|INFO|WARN|ERROR
  -tilesKey="": A key to protect your tiles access
```

## APIs

Tiles are available at `/tiles/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.pbf`, an optional `key` URL param can be passed to secure access to your tiles server, (use the `tilesKey` option).

Metrics are provided via Prometheus at `http://host:httpMetricsPort/metrics`.

A debug visual map is available at `http://host:httpAPIPort/static/`.

Health status is provided via gRPC `host:healthPort` or via HTTP `http://host:httpAPIPort/healthz`.

A `http://host:httpAPIPort/version` is giving you running version but also information on the dataset.

## Docker & Kubernetes

Main goal of kvtiles is to be run in Docker/Kubernetes with embedded maps.

A demo with Hawaii can be run:

```
 docker run --rm -it -p 8080:8080 akhenakh/kvtiles-demo:latest
```

Then point your browser to `http://yourdockerip:8080/static/`

An example deployment for kub is located in `cmd/kvtilesd`.
