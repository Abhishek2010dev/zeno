package main

import (
	"github.com/Abhishek2010dev/zeno"
)

func main() {
	z := zeno.New()
	z.Use(func(ctx *zeno.Context) error {
		return ctx.Next()
	})
	z.Get("/{id?}", func(ctx *zeno.Context) error {
		return ctx.SendString(ctx.Param("id"))
	})
	z.Run(":3000")
}
