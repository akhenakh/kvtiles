# kvtiles

kvtiles is a web server to embed and serve maps.

Using the MVT format extracted from MBTiles, kvtiles is using a key value storage to speed up queries.

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

## Docker & Kubernetes

Main goal of kvtiles is to be run in Docker/Kubernetes with embedded maps.

A demo with Hawaii can be run:

```
 docker run --rm -it -p 8080:8080 akhenakh/kvtiles-demo:latest
```

Then point your browser to http://yourdockerip:8080/static/

An example deployment for kub is located in `cmd/kvtilesd`.
