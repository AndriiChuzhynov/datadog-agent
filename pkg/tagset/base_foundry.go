package tagset

// baseFoundry provides some utility functions that are useful in all foundry
// implementations.
type baseFoundry struct {
	// builders is a cache of unused builder instances for reuse
	builders []*Builder
}

// newBuilder implements NewBuilder for a foundry
func (f *baseFoundry) newBuilder(ff Foundry, capacity int) *Builder {
	// NOTE: capacity is ignored (for now, pending later optimizations)
	var bldr *Builder
	if len(f.builders) > 0 {
		l := len(f.builders)
		bldr, f.builders = f.builders[l-1], f.builders[:l-1]
	} else {
		bldr = newBuilder(ff)
	}
	bldr.reset()
	return bldr
}

func (f *baseFoundry) builderClosed(builder *Builder) {
	f.builders = append(f.builders, builder)
}
