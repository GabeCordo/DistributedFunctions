package main

import (
	"fmt"

	"github.com/GabeCordo/DistributedFunctions"
	"github.com/GabeCordo/ScalingFunctions"
)

func main() {

	repository := ScalingFunctions.NewRepository()

	m := repository.Module("common")
	m.Version = "v1.0"
	err := m.LinkFunction("generator", generator)
	if err != nil {
		fmt.Print(err)
	}
	err = m.LinkFunction("add2", add2)
	if err != nil {
		fmt.Print(err)
	}
	err = m.LinkFunction("mul2", mul2)
	if err != nil {
		fmt.Print(err)
	}
	err = m.LinkFunction("prt", prt)
	if err != nil {
		fmt.Print(err)
	}

	processor := DistributedFunctions.New(repository)
	processor.Connect()

}
