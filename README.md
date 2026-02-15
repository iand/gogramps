# gogramps

Go package for reading and writing Gramps SQLite databases.

[![Go Report Card](https://goreportcard.com/badge/github.com/iand/gogramps)](https://goreportcard.com/report/github.com/iand/gogramps)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/iand/gogramps)

## About

This package provides Go types and a database API for reading and writing the native SQLite database format used by [Gramps](https://gramps-project.org/), an application for managing genealogical data.

It supports all 10 primary object types: Person, Family, Event, Place, Source, Citation, Repository, Note, Media, and Tag.

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
