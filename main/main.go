package main

import "github.com/ibrohimislam/luminescent"

func main() {

	coulomb := luminescent.Coulomb{}

	atom := luminescent.Atom{coulomb, luminescent.Proton{}}

	atom.Run()
}
