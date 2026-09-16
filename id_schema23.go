//go:build gramps_dbschema23

package gogramps

// The DNA object types that carry a Gramps ID.
const (
	ObjectDNATest  ObjectType = "dnatest"
	ObjectDNAMatch ObjectType = "dnamatch"
)

// init registers the default Gramps ID formats for the DNA object types.
func init() {
	defaultIDFormat[ObjectDNATest] = "T%05d"
	defaultIDFormat[ObjectDNAMatch] = "M%05d"
}
