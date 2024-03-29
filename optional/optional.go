package optional

import "errors"

var ErrNoneValue = errors.New("get None[T] value")

// Optional represents type that can either have value or be empty
type (
	Optional[T any] interface {
		// Get returns value of this Optional, or error
		Get() (T, error)
		// IsPresent tells whether instance holds value, in which case it returns true.
		IsPresent() bool
		// Filter returns this Optional if it's not empty and satisfies predicate. Otherwise it returns None
		Filter(p Predicate[T]) Optional[T]
		// OrElse returns this Optional if it's not empty, otherwise it returns other
		OrElse(other T) T
		// Unwrap - unwrap from Optional to usual T
		Unwrap() T
		// Error - unwrap from Optional error context if it exists.
		Error() error
	}

	// Value represents non-empty case of Optional
	Value[T any] struct {
		v T
		e error
	}

	// None - represents empty case of Optional
	None[T any] struct {
		e error
	}

	Func[T, V any]   func(T) V
	Func0[V any]     func() V
	Predicate[T any] func(T) bool
	Unit             struct{}
)

// New creates new Optional[T] from T
func New[T any](x T) Optional[T] {
	return Value[T]{v: x}
}

// NewP creates new Optional[T] from *T
func NewP[T any](x *T) Optional[T] {
	if x == nil {
		return None[T]{}
	}

	return Value[T]{v: *x}
}

// NewE creates new Optional[T] from T if err == nil, else create None with error context.
func NewE[T any](x T, err error) Optional[T] {
	if err != nil {
		return None[T]{err}
	}

	return New(x)
}

// NewPE creates new Optional[T] from *T if err == nil, else create None with err context.
func NewPE[T any](x *T, err error) Optional[T] {
	if err != nil {
		return None[T]{e: err}
	}

	return NewP(x)
}

// Empty creates None[T]
func Empty[T any]() Optional[T] {
	return None[T]{}
}

// EmptyE creates None[T] with set up error context
func EmptyE[T any](err error) Optional[T] {
	return None[T]{e: err}
}

func (j Value[T]) IsPresent() bool {
	return true
}

func (_ None[T]) IsPresent() bool {
	return false
}

func (j Value[T]) Filter(p Predicate[T]) Optional[T] {
	if p(j.v) {
		return j
	}

	return None[T]{}
}

// Filter on None returns None
func (n None[T]) Filter(_ Predicate[T]) Optional[T] {
	return n
}

func (j Value[T]) OrElse(_ T) T {
	return j.v
}

func (n None[T]) OrElse(other T) T {
	return other
}

func (j Value[T]) Get() (T, error) {
	return j.v, nil
}

func (n None[T]) Get() (T, error) {
	return *new(T), ErrNoneValue
}

func (j Value[T]) Unwrap() T {
	return j.v
}

func (n None[T]) Unwrap() T {
	return *new(T)
}

func (j Value[T]) Error() error {
	return j.e
}

func (n None[T]) Error() error {
	return n.e
}

// Map - Returns Value containing the result of applying f to m if it's non empty. Otherwise returns None
func Map[T, V any](m Optional[T], f Func[T, V]) Optional[V] {
	switch m.(type) {
	case Value[T]:
		return Value[V]{v: f(m.(Value[T]).v)}
	}
	return None[V]{}
}

// FlatMap - Same as Map but function must return Optional
func FlatMap[T, V any](m Optional[T], f Func[T, Optional[V]]) Optional[V] {
	switch m.(type) {
	case Value[T]:
		return f(m.(Value[T]).v)
	}
	return None[V]{}
}
