package main

import "github.com/ibrohimislam/amaryllis"

func main() {

	coulomb := amaryllis.Coulomb{}

	atom := amaryllis.Atom{coulomb, amaryllis.Proton{}}

	atom.Run()
}
