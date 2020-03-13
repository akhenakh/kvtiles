#!/bin/bash

MBTILESDIR=/opt/data/public_files

make mbtilestokv kvtilesd grpc_health_probe

mkdir -p build

while IFS=";" read -r name maxzoom file lat lng 
do
      echo "starting $name $maxzoom $file $lat $lng"
      rm -f "build/${name}.db"
      ./cmd/mbtilestokv/mbtilestokv -tilesPath ${MBTILESDIR}/${file} -dbPath "build/${name}.db" -maxZoom $maxzoom -centerLat $lat -centerLng $lng 
      cp "build/${name}.db" cmd/kvtilesd/map.db
      docker build cmd/kvtilesd -t akhenakh/kvtiles:${name}-latest
      docker push akhenakh/kvtiles:${name}-latest
      rm -f "build/${name}.db"
       
done < <(egrep -v "^#" maps.csv)


