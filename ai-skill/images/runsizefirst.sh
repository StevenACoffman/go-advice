#!/bin/bash

find . -name '*.svg' | while IFS= read -r svg; do
    base="${svg%.*}"
    for ext in png jpg jpeg PNG JPG JPEG; do
      raster="$base.$ext"
      if [ -f "$raster" ]; then
        svg_size=$(stat -f%z "$svg")
        raster_size=$(stat -f%z "$raster")
        if [ "$svg_size" -gt "$raster_size" ]; then
          printf "%d\t%d\t%s\n" "$svg_size" "$raster_size" "$svg"
        fi
        break
      fi
    done
  done | sort -rn > out.txt