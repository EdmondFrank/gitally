package git

const (
	// scReadOnly denotes a read-only command
	scReadOnly = 1 << iota
	// scNoRefUpdates denotes a command which will never update refs
	scNoRefUpdates
	// scNoEndOfOptions denotes a command which doesn't know --end-of-options
	scNoEndOfOptions
)

// subcommands is a curated list of Git command names for special git.SafeCmd
// validation logic
var subcommands = map[string]uint{
	"apply":            scNoRefUpdates,
	"archive":          scReadOnly | scNoEndOfOptions,
	"blame":            scReadOnly | scNoEndOfOptions,
	"bundle":           scReadOnly,
	"cat-file":         scReadOnly,
	"checkout":         scNoEndOfOptions,
	"clone":            scNoEndOfOptions,
	"commit-graph":     scNoRefUpdates,
	"config":           scNoRefUpdates | scNoEndOfOptions,
	"count-objects":    scReadOnly,
	"diff":             scReadOnly,
	"diff-tree":        scReadOnly,
	"for-each-ref":     scReadOnly | scNoEndOfOptions,
	"format-patch":     scReadOnly,
	"fsck":             scReadOnly,
	"gc":               scNoRefUpdates,
	"grep":             scReadOnly | scNoEndOfOptions,
	"hash-object":      scNoRefUpdates,
	"init":             scNoRefUpdates,
	"linguist":         scNoEndOfOptions,
	"log":              scReadOnly,
	"ls-remote":        scReadOnly,
	"ls-tree":          scReadOnly | scNoEndOfOptions,
	"merge-base":       scReadOnly,
	"multi-pack-index": scNoRefUpdates,
	"pack-refs":        scNoRefUpdates,
	"remote":           scNoEndOfOptions,
	"repack":           scNoRefUpdates,
	"rev-list":         scReadOnly,
	"rev-parse":        scReadOnly | scNoEndOfOptions,
	"show-ref":         scReadOnly,
	"upload-archive":   scReadOnly | scNoEndOfOptions,
	"upload-pack":      scReadOnly,
}

// mayUpdateRef indicates if a subcommand is known to update references.
// This is useful to determine if a command requires reference hook
// configuration. A non-exhaustive list of commands is consulted to determine if
// refs are updated. When unknown, true is returned to err on the side of
// caution.
func mayUpdateRef(subcmd string) bool {
	flags, ok := subcommands[subcmd]
	if !ok {
		return true
	}
	return flags&(scReadOnly|scNoRefUpdates) == 0
}

// supportsEndOfOptions indicates whether a command can handle the
// `--end-of-options` option.
func supportsEndOfOptions(subcmd string) bool {
	flags, ok := subcommands[subcmd]
	if !ok {
		return true
	}
	return flags&scNoEndOfOptions == 0
}
