# gogramps

Go package for reading and writing Gramps SQLite databases.

[![Go Report Card](https://goreportcard.com/badge/github.com/iand/gogramps)](https://goreportcard.com/report/github.com/iand/gogramps)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/iand/gogramps)

## About

This package provides Go types and a database API for reading and writing the native SQLite database format used by [Gramps](https://gramps-project.org/), an application for managing genealogical data.

It supports all 10 primary object types: Person, Family, Event, Place, Source, Citation, Repository, Note, Media, and Tag. Schema versions 21 and 22 are supported.

Experimental support for schema 23 is available using the `gramps_schema23` build tag. Schema 23 adds the DNATest and DNAMatch object types, which are being developed in the upstream Gramps project (see the [design discussion](https://github.com/gramps-project/gramps/discussions/2292) and [implementation pull request](https://github.com/gramps-project/gramps/pull/2295)). The schema 23 API may change as the upstream design evolves.

```sh
go build -tags gramps_schema23 ./...
go test -tags gramps_schema23 ./...
```

## Status

This is an early release. Use with caution.

## Usage

```Go
package main

import (
	"fmt"
	"log"

	"github.com/iand/gogramps"
)

func main() {
	// Open an existing Gramps database directory.
	db, err := gogramps.Open("/path/to/gramps/database")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Iterate over all people in the database.
	for person, err := range db.People() {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s %s\n",
			person.GrampsID,
			person.PrimaryName.FirstName,
			person.PrimaryName.SurnameList[0].Surname,
		)
	}
}
```

### Creating a new database

```Go
db, err := gogramps.Create("/path/to/new/database", "My Family Tree")
if err != nil {
	log.Fatal(err)
}
defer db.Close()

person := &gogramps.Person{
	Handle:   gogramps.NewHandle(),
	GrampsID: "I0001",
	Gender:   gogramps.GenderMale,
	PrimaryName: gogramps.Name{
		Class:     "Name",
		FirstName: "John",
		SurnameList: []gogramps.Surname{
			{Class: "Surname", Surname: "Doe", Primary: true},
		},
		Type: gogramps.GrampsType{Class: "NameType"},
	},
}

if err := db.AddPerson(person); err != nil {
	log.Fatal(err)
}
```

## Getting Started

Run the following in the directory containing your project's `go.mod` file:

```sh
go get github.com/iand/gogramps@latest
```

Documentation is at [https://pkg.go.dev/github.com/iand/gogramps](https://pkg.go.dev/github.com/iand/gogramps)

## License

This is free software released under the GNU General Public License v2.0. See the accompanying [`COPYING`](COPYING) file for details.
