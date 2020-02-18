<p align="center">
  <img src="https://hsto.org/webt/ek/0q/co/ek0qcotut4ioaifdup8nroyko2a.png" alt="logo" width="128" />
</p>

# `go-filecache`

![Release version][badge_release_version]
![Project language][badge_language]
[![Build Status][badge_build]][link_build]
[![Coverage][badge_coverage]][link_coverage]
[![Go Report][badge_goreport]][link_goreport]
[![License][badge_license]][link_license]

This package provides file-based cache implementation with entries expiration, checksum validation and another useful features.

Methods that interacts with file system uses mutexes, so, this cache implementation thread-safe, but performance in this case can be less than you might want.

## Installation and usage

The import path for the package is `github.com/tarampampam/go-filecache`.

To install it, run:

```bash
go get github.com/tarampampam/go-filecache
```

> API documentation can be [found here](https://godoc.org/github.com/tarampampam/go-filecache).

### Usage example

```go
package main

import (
    "bytes"
    "fmt"
    "time"

    filecache "github.com/tarampampam/go-filecache"
)

func main() {
    // Create new cache items pool
	// Note: Do NOT create new pool instance in goroutine - SHARE it instead
    pool := filecache.NewPool("/tmp")
    
    // Put data into cache pool with expiration time
    expiresAt := time.Now().Add(time.Minute * 10)
    if _, err := pool.Put("foo", bytes.NewBuffer([]byte("foo data")), expiresAt); err != nil {
        panic(err)
    }
    
    // Put data without expiration time
    if _, err := pool.PutForever("bar", bytes.NewBuffer([]byte("bar data"))); err != nil {
        panic(err)
    }

    // Define buffer for cached data reading
    buf := bytes.NewBuffer([]byte{})

    // Read data using reader
    if err := pool.GetItem("foo").Get(buf); err != nil {
        panic(err)
    }

    fmt.Println(buf) // "foo data"
}
```

Simple benchmark _(put value, get value, read into buffer and set new value)_ results:

```
goos: linux
goarch: amd64
pkg: github.com/tarampampam/go-filecache
BenchmarkSetAndGet-8       10000            611470 ns/op           40836 B/op         84 allocs/op
```

### Testing

For application testing we use built-in golang testing feature and `docker-ce` + `docker-compose` as develop environment. So, just write into your terminal after repository cloning:

```shell
$ make test
```

Or execute benchmarks:

```shell
$ make gobench
```

## Changelog

[![Release date][badge_release_date]][link_releases]
[![Commits since latest release][badge_commits_since_release]][link_commits]

Changes log can be [found here][link_changes_log].

## Support

[![Issues][badge_issues]][link_issues]
[![Issues][badge_pulls]][link_pulls]

If you will find any package errors, please, [make an issue][link_create_issue] in current repository.

## License

This is open-sourced software licensed under the [MIT License][link_license].

[badge_build]:https://img.shields.io/github/workflow/status/tarampampam/go-filecache/build?maxAge=30&logo=github
[badge_coverage]:https://img.shields.io/codecov/c/github/tarampampam/go-filecache/master.svg?maxAge=30
[badge_goreport]:https://goreportcard.com/badge/github.com/tarampampam/go-filecache
[badge_size_latest]:https://images.microbadger.com/badges/image/tarampampam/go-filecache.svg
[badge_release_version]:https://img.shields.io/github/release/tarampampam/go-filecache.svg?maxAge=30
[badge_language]:https://img.shields.io/github/go-mod/go-version/tarampampam/go-filecache?longCache=true
[badge_license]:https://img.shields.io/github/license/tarampampam/go-filecache.svg?longCache=true
[badge_release_date]:https://img.shields.io/github/release-date/tarampampam/go-filecache.svg?maxAge=180
[badge_commits_since_release]:https://img.shields.io/github/commits-since/tarampampam/go-filecache/latest.svg?maxAge=45
[badge_issues]:https://img.shields.io/github/issues/tarampampam/go-filecache.svg?maxAge=45
[badge_pulls]:https://img.shields.io/github/issues-pr/tarampampam/go-filecache.svg?maxAge=45
[link_goreport]:https://goreportcard.com/report/github.com/tarampampam/go-filecache

[link_coverage]:https://codecov.io/gh/tarampampam/go-filecache
[link_build]:https://github.com/tarampampam/go-filecache/actions
[link_license]:https://github.com/tarampampam/go-filecache/blob/master/LICENSE
[link_releases]:https://github.com/tarampampam/go-filecache/releases
[link_commits]:https://github.com/tarampampam/go-filecache/commits
[link_changes_log]:https://github.com/tarampampam/go-filecache/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/go-filecache/issues
[link_create_issue]:https://github.com/tarampampam/go-filecache/issues/new/choose
[link_pulls]:https://github.com/tarampampam/go-filecache/pulls
