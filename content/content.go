// Package content embeds the default portfolio content into the binary, so
// the server runs with no extra files. Override it at runtime with
// `-content path/to/portfolio.yaml`.
package content

import _ "embed"

//go:embed portfolio.yaml
var Default []byte
