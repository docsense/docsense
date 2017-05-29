package main

import (
	do "gopkg.in/godo.v2"
)

func tasks(p *do.Project) {
	//do.Env = `GOPATH=.vendor::$GOPATH`

	p.Task("default", do.S{"run"}, nil)

	p.Task("build", do.S{"run"}, func(c *do.Context) {
		c.Run("GOOS=linux GOARCH=amd64 go build", do.M{"$in": "./"})
	}).Src("**/*.go")

	p.Task("run", do.S{}, func(c *do.Context) {
		c.Start("main.go", do.M{"$in": "./core"})
	}).Src("*.go", "disk/*.go", "disk/plugins/*.go", "mongo/*.go", "disk/plugins/s3/*.go", "search/*.go", "search/index/*.go", "common/*.go", "core/*.go", "search/fulltext/*.go", "metadata/*.go").
		Debounce(3000)
}

func main() {
	do.Godo(tasks)
}
