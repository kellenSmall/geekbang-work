package ExerciseUnsafe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type User struct {
	Age int
}

func TestUnsafeAccessor_Field(t *testing.T) {
	testCase := []struct {
		name     string
		inputVal any
		field    string
		newVal   any

		wantVal int
		wantErr error
	}{
		{
			name:     "test nil",
			inputVal: nil,
			wantErr:  errInvalidEntity,
		},
		{
			name:     "test nil",
			inputVal: User{Age: 1},
			wantErr:  errInvalidEntity,
		},
		{
			name: "test User",
			inputVal: &User{
				Age: 18,
			},
			field:   "Age",
			wantVal: 18,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {

			//"github.com/stretchr/testify/assert"

			unsafeAccessor, err := NewUnsafeAccessor(tc.inputVal)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			val, err := unsafeAccessor.Field(tc.field)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func TestUnsafeAccessor_SetField(t *testing.T) {

	testCase := []struct {
		name     string
		inputVal *User
		field    string
		newVal   int

		wantVal int
		wantErr error
	}{
		{
			name:     "test nil",
			inputVal: nil,
			wantErr:  errInvalidEntity,
		},
		{
			name: "test User",
			inputVal: &User{
				Age: 18,
			},
			newVal:  20,
			field:   "Age",
			wantVal: 20,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {

			//"github.com/stretchr/testify/assert"

			unsafeAccessor, err := NewUnsafeAccessor(tc.inputVal)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			err = unsafeAccessor.SetField(tc.field, tc.newVal)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.newVal, tc.inputVal.Age)
		})
	}

}
