#!/bin/bash

root=$(cd $(dirname $0) && cd .. && pwd)

cat > example/config.yml <<EOF
resource_map:
  some-resource: $root/example/some-resource
EOF
