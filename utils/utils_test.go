package utils

import (
	"reflect"
	"testing"
)

func TestListSelect2(t *testing.T) {
	type args struct {
		list []string
		max  int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "Test 1",
			args: args{
				list: []string{"1", "4-6", "8"},
				max:  10,
			},
			want: []int{1, 4, 5, 6, 8},
		},
		{
			name: "items test",
			args: args{
				max:  3,
				list: []string{"1", "3"},
			},
			want: []int{1, 3},
		},
		{
			name: "from to item selection 1",
			args: args{
				max:  10,
				list: []string{"1-3", "5", "7-8", "10"},
			},
			want: []int{1, 2, 3, 5, 7, 8, 10},
		},
		{
			name: "from to item selection 2",
			args: args{
				max:  10,
				list: []string{"1", "2", "4", "5", "7-8", "10"},
			},
			want: []int{1, 2, 4, 5, 7, 8, 10},
		},
		{
			name: "from to item selection 3",
			args: args{
				max:  10,
				list: []string{"5-1", "2"},
			},
			want: []int{2},
		},
		{
			name: "from to item selection 4",
			args: args{
				max:  10,
				list: []string{"5", "8-10"},
			},
			want: []int{5, 8, 9, 10},
		},
		{
			name: "from to item selection 5",
			args: args{
				max:  6,
				list: []string{"1-3", "8-10"},
			},
			want: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListSelect(tt.args.list, tt.args.max); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}
