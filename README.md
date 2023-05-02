# TinyKV

[![build](https://github.com/micpst/tinykv/actions/workflows/build.yml/badge.svg)](https://github.com/micpst/tinykv/actions/workflows/build.yml)

Distributed key-value store written in Go.

## 🛠️ Installation
### Build from source
To build and run master server from the source code:
1. Requirements: **go, make & nginx**
2. Install dependencies:
```bash
$ make setup
```
3. Build:
```bash
$ make build
```
4. Setup volume instances:
```bash
$ PORT=3001 VOLUME=tmp/vol1 ./volume/setup.sh
$ PORT=3002 VOLUME=tmp/vol2 ./volume/setup.sh
$ PORT=3003 VOLUME=tmp/vol3 ./volume/setup.sh
```
5. Run the master server binary:
```bash
$ ./bin/master --db ./tmp/indexdb/ --port 3000 --volumes localhost:3001,localhost:3002,localhost:3003
```

## 📘 Usage
### Master API
Put `"bigswag"` in `"wehave"` key:
```bash
$ curl -L -X PUT -d bigswag localhost:3000/wehave
```

Get `"wehave"` key:
```bash
$ curl -L localhost:3000/wehave
```

Delete `"wehave"` key:
```bash
$ curl -L -X DELETE localhost:3000/wehave
```

List keys starting with `"we"`:
```bash
$ curl -L "localhost:3000/we?list"
$ curl -L "localhost:3000/we?list&limit=100"
$ curl -L "localhost:3000?list&start=/we&limit=100"
```

### Rebalance volumes
Change the amount of volume servers:
```bash
$ ./bin/master --cmd rebalance --db ./tmp/indexdb/ --volumes localhost:3001,localhost:3002
```
> Before rebalancing, make sure the master server is down, as LevelDB can only be accessed by one process.

### Rebuild the index
Regenerate the LevelDB:
```bash
$ ./bin/master --cmd rebuild --db ./tmp/indexdb-alt/ --volumes localhost:3001,localhost:3002,localhost:3003
```

## 🕜 Performance
Fetching non-existent key: ~10325 req/sec
```bash
$ wrk -t2 -c100 -d10s http://localhost:3000/key
```
Fetching existent key: ~9800 req/sec
```bash
$ wrk -t2 -c100 -d10s http://localhost:3000/wehave
```

## 📄 License
All my code is MIT licensed. Libraries follow their respective licenses.
