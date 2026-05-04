#!/bin/bash
find . \
    -type f \( -iname '*.png' -o -iname '*.jpg' -o -iname '*.jpeg' \) |
  while IFS= read -r img; do
    out="${img%.*}.svg"
    echo "Converting: $img → $out"
    vtracer --input "$img" --output "$out"
  done