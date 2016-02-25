[![GoDoc](https://godoc.org/gopkg.in/paperstreet/json.v0?status.svg)](https://godoc.org/gopkg.in/paperstreet/json.v0)
[![Build Status](https://travis-ci.org/paperstreet/json.svg?branch=master)](https://travis-ci.org/paperstreet/json)

# json
Streaming json encoding in Go. I'm still not completely happy with the api, so
it may change. Import using `"gopkg.in/paperstreet/json.v0"`.

    BenchmarkStdlib-4       1000     1734826 ns/op     5.64 MB/s    232768 B/op     3080 allocs/op
    BenchmarkBuilder-4      1000     1629830 ns/op     6.00 MB/s    416133 B/op     7002 allocs/op
