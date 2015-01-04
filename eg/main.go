package main

import (
	"fmt"
	"github.com/echlebek/args"
	"os"
)

type Args struct {
	Foo int     `args:"this is a foo,-f"` // Foo has a short flag, -f
	Bar float32 `args:"this is a bar,r"`  // Bar is required
	Baz string  `args:"a baz!,r"`         // Baz is required too
}

// with args, set defaults by putting the default you want in the struct.
var defaultArgs = Args{Foo: 5}

func main() {
	a := defaultArgs

	if err := args.Parse(&a); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		if err := args.Usage(os.Stderr, defaultArgs); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		return
	}

	fmt.Printf("%+v\n", a)
}
