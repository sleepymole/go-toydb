module github.com/sleepymole/go-toydb

go 1.22

require (
	github.com/emirpasic/gods/v2 v2.0.0-alpha
	github.com/google/btree v1.1.2
	github.com/google/uuid v1.6.0
	github.com/samber/lo v1.39.0
	github.com/samber/mo v1.11.0
	google.golang.org/protobuf v1.33.0
)

require golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect

replace github.com/emirpasic/gods/v2 => github.com/sleepymole/gods/v2 v2.0.0-20240308045935-7c91a38e2a2e
