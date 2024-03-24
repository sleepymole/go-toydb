module github.com/sleepymole/go-toydb

go 1.22

require (
	github.com/emirpasic/gods/v2 v2.0.0-alpha
	github.com/google/btree v1.1.2
	github.com/google/uuid v1.6.0
	github.com/samber/lo v1.39.0
	github.com/samber/mo v1.11.0
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
)

replace github.com/emirpasic/gods/v2 => github.com/sleepymole/gods/v2 v2.0.0-20240308045935-7c91a38e2a2e
