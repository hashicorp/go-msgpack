# go

Collection of Open-Source Go libraries and tools.

## v2 versioning
Upstream `github.com/ugorji/go` made several breaking changes that led to HashiCorp maintaining its own fork.
Initially, tag `v1.1.5` contained these up-to-date changes from upstream but that caused issues when projects 
depending on `v0.5.5` behavior imported modules depending on `v1.1.5`. By retracting `v1.1.5` and re-releasing 
it as `v2.0.0` using go's [v2](https://go.dev/blog/v2-go-modules) semantics, projects can indirectly import 
both dependencies.

## Codec

[Codec](https://github.com/ugorji/go/tree/master/codec#readme) is a High Performance and Feature-Rich Idiomatic encode/decode and rpc library for [msgpack](http://msgpack.org) and [Binc](https://github.com/ugorji/binc).

Online documentation is at [http://godoc.org/github.com/ugorji/go/codec].

Install using:

    go get github.com/ugorji/go/codec

