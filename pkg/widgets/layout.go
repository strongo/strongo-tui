package widgets

// Size is the share of a length one part of a layout takes: a fixed number of
// cells, or a weight in what remains.
type Size struct {
	fixed  int
	weight int
}

// Fixed is a part of exactly n cells.
func Fixed(n int) Size { return Size{fixed: max(n, 0)} }

// Fill is a part that takes a share of the space left after the fixed parts,
// proportional to weight.
func Fill(weight int) Size { return Size{weight: max(weight, 0)} }

// Split divides total cells among parts. Fixed parts are served first, in order,
// and are cut short when they do not fit; the rest is shared among Fill parts in
// proportion to their weights, the last of them taking the rounding remainder.
// A Fill part with weight zero gets nothing.
func Split(total int, parts ...Size) []int {
	sizes := make([]int, len(parts))
	remaining := max(total, 0)
	weight := 0
	for i, p := range parts {
		if p.fixed > 0 {
			sizes[i] = min(p.fixed, remaining)
			remaining -= sizes[i]
		} else {
			weight += p.weight
		}
	}
	if weight == 0 {
		return sizes
	}
	share := remaining
	last := -1
	for i, p := range parts {
		if p.fixed > 0 || p.weight == 0 {
			continue
		}
		sizes[i] = share * p.weight / weight
		remaining -= sizes[i]
		last = i
	}
	sizes[last] += remaining
	return sizes
}
