# muss

Definition:

> noun: mess; scramble; a confused conflict
> verb: to put into disorder; to make messy

Or if you prefer, an acronym:

> Manage User Selected Services

`muss` is a command line application to allow users to choose what services
they want to run and how they want to run them.

Projects (and their dependent services) can be configured
in a familiar style through a series of yaml files.

More info to come...

# Building, Testing, etc

`muss` is a CLI written in go and can be built with the usual go commands:

`go install` to install it into your `$GOPATH`

`go test -v ./...` to test all of the included go packages.
