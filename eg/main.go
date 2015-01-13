package main

import (
	"fmt"
	"github.com/echlebek/args"
	"os"
)

// Args is the spec for the arguments we want our program to accept.
type Args struct {
	Foo int     `args:"this is a foo,-f"` // Foo has a short flag, -f
	Bar float32 `args:"this is a bar,r"`  // Bar is required
	Baz string  `args:"a baz!,r"`         // Baz is required too
}

// With args, users set defaults by putting the default they want in the struct.
// If a value type is used, the default will be the zero value.
// If a pointer type is used, there will be no default.
var defaultArgs = Args{Foo: 5}

func main() {
	a := defaultArgs // copy defaultArgs

	// Try to parse args into a. Internally, args reads os.Args.
	if err := args.Parse(&a); err != nil {
		// Will print an error if an argument is malformed.
		fmt.Fprintf(os.Stderr, "%s\n", err)
		// Print the usage
		if err := args.Usage(os.Stderr, defaultArgs); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		return
	}

	fmt.Printf("%+v\n", a) // Print out the args we received.
}
