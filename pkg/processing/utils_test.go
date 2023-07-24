package processing

import "testing"

func Test_slicesEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want bool
	}{
		{
			name: "identical slices",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: true,
		},
		{
			name: "different order",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "b", "a"},
			want: true,
		},
		{
			name: "different lengths",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slicesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("slicesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
