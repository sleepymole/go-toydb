package protoutil

import "google.golang.org/protobuf/proto"

func Clone[T proto.Message](msg T) T {
	clone := proto.Clone(msg)
	return clone.(T)
}

func MustMarshal[T proto.Message](msg T) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return data
}
