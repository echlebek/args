args
====

An experimental argument parser for go.

This is a small project that I used to teach myself the details of Go's reflect package.
While I have done other work with Go's reflect package, I wanted a sandbox project
for experimenting with the reflect package's techniques.

The basic idea here is to create a struct that has the arguments you wish to accept
for a command-line program. args.Parse() then takes the struct as an argument, and
tries to match arguments from the command line.

Unlike the flag package, args is a GNU-style argument parser.

Errors try to be sensible but are not well tested.

Some options can be supplied via struct tags. (See eg/main.go)

Documentation: https://godoc.org/github.com/echlebek/args
