// tools project main.go
package main

import (
	"flag"
	"tools/cmd"
)

func main() {
	f := flag.Bool("f", false, "force action ( true | false )")
	show := flag.Bool("show", false, "only show not action ( true | false )")
	env := flag.String("env", "all", "operate env ( all | test | stage | product )")
	flag.Parse()

	args := flag.Args()

	for _, arg := range args {
		cmd.FindDeploy(arg, *f, *show, *env)
		cmd.RmMongo(arg, *f, *show, *env)
	}

}
