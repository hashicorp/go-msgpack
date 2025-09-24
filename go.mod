module github.com/hashicorp/go-msgpack/v2

go 1.24.0

toolchain go1.24.7

retract v2.1.4 // Contains unnecessarily high go 1.25.1 build requirement

require (
	github.com/Sereal/Sereal/Go/sereal v0.0.0-20231009093132-b9187f1a92c6
	github.com/davecgh/go-xdr v0.0.0-20161123171359-e6a2ba005892
	github.com/json-iterator/go v1.1.12
	github.com/mailru/easyjson v0.7.7
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/tinylib/msgp v1.1.8
	golang.org/x/tools v0.37.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/vmihailenco/msgpack.v2 v2.9.2
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
