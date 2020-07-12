package main

import (
	"context"
	"github.com/RaniSputnik/lovepac/packer"
	"github.com/RaniSputnik/lovepac/target"
	"log"
)

func main() {
	params := packer.Params{
		Name:   "atlas",
		Format: target.Starling,
		Input:  packer.NewFileStream("./assets/texture"),
		Output: packer.NewFileOutputter("./assets"),
		Width:  512,
		Height: 512,
	}
	var err error

	err = packer.Run(context.Background(), &params)
	if err != nil {
		log.Fatal(err)
	}
}
