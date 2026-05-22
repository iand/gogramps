//go:build !gramps_schema23

package gogramps

const maxSupportedSchemaVersion = 22

func schema23Tables() []string { return nil }
