package tagset

// DefaultFoundry is a global thread-safe foundry, used by calls to
// package-level functions.  This is suitable for non-performance-critical tags
// manipulation
var DefaultFoundry Foundry

func init() {
	DefaultFoundry = NewThreadsafeFoundry(newCachingFoundry())
}

// NewTags calls DefaultFoundry.NewTags
func NewTags(tags ...string) *Tags {
	return DefaultFoundry.NewTags(tags...)
}

// NewUniqueTags calls DefaultFoundry.NewUniqueTags
func NewUniqueTags(tags ...string) *Tags {
	return DefaultFoundry.NewUniqueTags(tags...)
}

// NewTagsFromMap calls DefaultFoundry.NewTagsFromMap
func NewTagsFromMap(tags map[string]struct{}) *Tags {
	return DefaultFoundry.NewTagsFromMap(tags)
}

// NewBuilder calls DefaultFoundry.NewBuilder
func NewBuilder(capacity int) *Builder {
	return DefaultFoundry.NewBuilder(capacity)
}

// TODO: NewSliceBuilder
// TODO: UnmarshalJSON
// TODO: UnmarshalYAML

// ParseDSD calls DefaultFoundry.ParseDSD
func ParseDSD(data []byte) (*Tags, error) {
	return DefaultFoundry.ParseDSD(data)
}

// Union calls DefaultFoundry.Union
func Union(a, b *Tags) *Tags {
	return DefaultFoundry.Union(a, b)
}

// DisjointUnion calls DefaultFoundry.DisjointUnion
func DisjointUnion(a, b *Tags) *Tags {
	return DefaultFoundry.DisjointUnion(a, b)
}
