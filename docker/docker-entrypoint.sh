#!/busybox/sh

./master -cmd rebuild -db indexdb -volumes "$VOLUMES"
./master -cmd rebalance -db indexdb -volumes "$VOLUMES" -replicas "$REPLICAS"
./master -cmd run -p 3000 -db indexdb -volumes "$VOLUMES" -replicas "$REPLICAS"
