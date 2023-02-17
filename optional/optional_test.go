package optional

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	assert.NotNil(t, New(struct{}{}))
	assert.NotNil(t, New(&struct{}{}))
	assert.NotNil(t, NewP(&struct{}{}))
	assert.NotNil(t, NewE(struct{}{}, nil))
	assert.NotNil(t, NewE(struct{}{}, ErrNoneValue))
	assert.NotNil(t, NewPE(&struct{}{}, nil))
	assert.NotNil(t, NewPE(&struct{}{}, ErrNoneValue))
}

type S struct{ v int }

func Test_New_Methods(t *testing.T) {
	var (
		a = S{v: 10}
		b = S{v: 10}
	)

	var opt = New(a)

	assert.NotNil(t, opt)
	v, err := opt.Get()
	assert.Equal(t, a, v)
	assert.Nil(t, err)
	assert.True(t, opt.IsPresent())

	v = opt.OrElse(b)
	assert.Equal(t, a, v)

	var opt2 = New(&a)

	assert.NotNil(t, opt2)
	v2, err := opt2.Get()
	assert.Equal(t, &a, v2)
	assert.Nil(t, err)
	assert.True(t, opt2.IsPresent())

	var opt3 = Empty[S]()

	assert.NotNil(t, opt3)
	v3, err := opt3.Get()
	assert.Equal(t, S{}, v3)
	assert.Equal(t, ErrNoneValue, err)
	assert.False(t, opt3.IsPresent())

	v3 = opt3.OrElse(S{})
	assert.Equal(t, S{}, v3)
}

func Test_Map(t *testing.T) {
	opt := New(S{v: 10})
	mult2 := func(x S) S { return S{v: x.v * 2} }

	v, err := Map(opt, mult2).Get()
	assert.Equal(t, 20, v.v)
	assert.Nil(t, err)

	opt = Empty[S]()

	v, err = Map(opt, mult2).Get()
	assert.Equal(t, ErrNoneValue, err)
	assert.Equal(t, 0, v.v)

}

func Test_FlatMap(t *testing.T) {
	opt := New(S{v: 10})

	mult2 := func(x S) Optional[S] { return New(S{v: x.v * 2}) }

	v, err := FlatMap(opt, mult2).Get()
	assert.Equal(t, 20, v.v)
	assert.Nil(t, err)

	opt = Empty[S]()

	v, err = FlatMap(opt, mult2).Get()
	assert.Equal(t, ErrNoneValue, err)
	assert.Equal(t, 0, v.v)
}

func Test_Predicates(t *testing.T) {
	opt := New(S{v: 10})

	gtzero := func(x S) bool { return x.v > 0 }

	assert.True(t, opt.Filter(gtzero).IsPresent())

	opt = New(S{v: -1})

	assert.False(t, opt.Filter(gtzero).IsPresent())

	opt = Empty[S]()

	assert.False(t, opt.Filter(gtzero).IsPresent())
}

func Test_Unwrap_Methods(t *testing.T) {
	obj := S{v: 10}

	assert.Equal(t, obj, New(obj).Unwrap())
	assert.Equal(t, obj, NewP(&obj).Unwrap())

	assert.Equal(t, S{}, None[S]{}.Unwrap())
}

func Test_Error_Methods(t *testing.T) {
	obj := S{v: 10}

	assert.NoError(t, New(obj).Error())
	err := errors.New("123")
	assert.Equal(t, err, NewE(obj, err).Error())

	assert.NoError(t, Empty[S]().Error())
	assert.Equal(t, err, EmptyE[S](err).Error())
}
