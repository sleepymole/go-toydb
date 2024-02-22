package rangeutil

import "bytes"

type Bound struct {
	value    []byte
	included bool
	excluded bool
}

func (b Bound) Value() []byte {
	return b.value
}

func (b Bound) Included() bool {
	return b.included
}

func (b Bound) Excluded() bool {
	return b.excluded
}

func (b Bound) Unbounded() bool {
	return !b.included && !b.excluded
}

func (b Bound) Clone() Bound {
	return Bound{
		value:    bytes.Clone(b.value),
		included: b.included,
		excluded: b.excluded,
	}
}

func Included(value []byte) Bound {
	return Bound{value: value, included: true}
}

func Excluded(value []byte) Bound {
	return Bound{value: value, excluded: true}
}

var Unbounded = Bound{}

func Range(start, end []byte) (Bound, Bound) {
	return Included(start), Excluded(end)
}

func RangeFull() (Bound, Bound) {
	return Unbounded, Unbounded
}

func RangeInclusive(start, end []byte) (Bound, Bound) {
	return Included(start), Included(end)
}

func RangeFrom(start []byte) (Bound, Bound) {
	return Included(start), Unbounded
}

func RangeTo(end []byte) (Bound, Bound) {
	return Unbounded, Excluded(end)
}

func RangeToInclusive(end []byte) (Bound, Bound) {
	return Unbounded, Included(end)
}
