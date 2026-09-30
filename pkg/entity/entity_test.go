package entity

import "testing"

func TestRefSame(t *testing.T) {
	a := Ref{Type: "t", Keys: map[string]string{"id": "1"}, Title: "A"}
	cases := []struct {
		name string
		b    Ref
		want bool
	}{
		{"same ignoring title", Ref{Type: "t", Keys: map[string]string{"id": "1"}, Title: "B"}, true},
		{"different type", Ref{Type: "u", Keys: map[string]string{"id": "1"}}, false},
		{"different key count", Ref{Type: "t", Keys: map[string]string{"id": "1", "x": "y"}}, false},
		{"different value", Ref{Type: "t", Keys: map[string]string{"id": "2"}}, false},
		{"different key name", Ref{Type: "t", Keys: map[string]string{"other": "1"}}, false},
	}
	for _, c := range cases {
		if got := a.Same(c.b); got != c.want {
			t.Errorf("%s: Same = %v, want %v", c.name, got, c.want)
		}
	}
	if !(Ref{Type: "t"}).Same(Ref{Type: "t"}) {
		t.Error("refs without keys must be Same")
	}
}
