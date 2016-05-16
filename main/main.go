package main

import "github.com/ibrohimislam/luminescent"
import "flag"

var port *int

func init() {
	port = flag.Int("port", 8080, "Description")
}

func main() {
	flag.Parse()

	coulomb := luminescent.Coulomb{}

	atom := luminescent.Atom{coulomb, luminescent.Proton{}}

	atom.Run(*port)
}
