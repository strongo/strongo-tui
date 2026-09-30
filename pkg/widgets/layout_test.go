package widgets_test

import (
	"slices"
	"testing"

	"github.com/tuigoff/tuigoff/pkg/widgets"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name  string
		total int
		parts []widgets.Size
		want  []int
	}{
		{"fixed and fill", 10, []widgets.Size{widgets.Fixed(1), widgets.Fill(1), widgets.Fixed(2)}, []int{1, 7, 2}},
		{"proportions with remainder", 10, []widgets.Size{widgets.Fill(1), widgets.Fill(2)}, []int{3, 7}},
		{"fixed cut short", 4, []widgets.Size{widgets.Fixed(10), widgets.Fixed(5)}, []int{4, 0}},
		{"no fill leaves a gap", 10, []widgets.Size{widgets.Fixed(3)}, []int{3}},
		{"zero weight gets nothing", 5, []widgets.Size{widgets.Fill(0), widgets.Fill(1)}, []int{0, 5}},
		{"only zero weights", 5, []widgets.Size{widgets.Fill(0)}, []int{0}},
		{"negative sizes clamp", -3, []widgets.Size{widgets.Fixed(-1), widgets.Fill(-2)}, []int{0, 0}},
		{"nothing", 5, nil, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := widgets.Split(tt.total, tt.parts...); !slices.Equal(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
