package main

import "github.com/Abhishek2010dev/zeno"

func main() {
	z := zeno.New()
	z.Get("/", func(ctx *zeno.Context) error {
		return ctx.SendJSON(map[string]string{"message": "Hello, World"})
	})
	z.Run(":3000")
}
