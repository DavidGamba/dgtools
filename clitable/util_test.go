package clitable

import (
	"reflect"
	"testing"
)

func TestPrefixSlice(t *testing.T) {
	type args struct {
		prefix []string
		base   []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple",
			args{prefix: []string{"b", "c"}, base: []string{"a", "b", "c"}},
			[]string{"b", "c", "a"},
			false,
		},
		{
			"drop extra",
			args{prefix: []string{"d", "c"}, base: []string{"a", "b", "c"}},
			[]string{"c", "a", "b"},
			false,
		},
		{
			"empty",
			args{prefix: []string{}, base: []string{"a", "b", "c"}},
			[]string{"a", "b", "c"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prefixSlice(tt.args.prefix, tt.args.base)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prefixSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
