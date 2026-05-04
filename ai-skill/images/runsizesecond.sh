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
  done | sort -rn | awk '
  BEGIN { printf "%-8s  %-8s  %-5s  %s\n", "SVG", "RASTER", "BLOAT", "FILE"; print "" }
  {
    svg_mb   = $1 / 1048576
    rast_mb  = $2 / 1048576
    ratio    = $1 / $2
    path     = $3
    sub(".*/images/", "", path)
    printf "%-8s  %-8s  %4.1fx  %s\n",
      sprintf("%.1fMB", svg_mb),
      sprintf("%.1fMB", rast_mb),
      ratio, path
    total_svg  += $1
    total_rast += $2
    count++
  }
  END {
    print ""
    printf "%d files, %.0fMB wasted (SVG total %.0fMB vs raster total %.0fMB)\n",
      count, (total_svg - total_rast) / 1048576,
      total_svg / 1048576, total_rast / 1048576
  }'
