package codec

type vocab map[string]uint

// reverse maps token IDs to tokens. IDs are dense, so a slice indexed by ID
// replaces a map; an empty string marks an ID with no token.
type reverse []string
