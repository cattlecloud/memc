package memc

import (
	"testing"

	"github.com/shoenig/test/must"
)

func Test_encode(t *testing.T) {
	t.Run("[]byte", func(t *testing.T) {
		b, err := encode([]byte{2, 4, 6, 8})
		must.NoError(t, err)
		must.SliceLen(t, 4, b)
	})

	t.Run("string", func(t *testing.T) {
		b, err := encode("foobar")
		must.NoError(t, err)
		must.SliceLen(t, 6, b)
	})

	t.Run("int8", func(t *testing.T) {
		var i int8 = 3
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 1, b)
	})

	t.Run("uint8", func(t *testing.T) {
		var i uint8 = 3
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 1, b)
	})

	// TODO all the types
}

func Test_decode(t *testing.T) {
	// TODO
}
