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
5. Run the server binary:
```bash
$ ./bin/server
```

## 📘 Usage
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
