package dcesecurity

// Domain represents the identifier for a local domain
type Domain byte

const (
	User  Domain = iota + 1 // POSIX UID domain
	Group                   // POSIX GID domain
)
