package main

import (
  "log"
  "os"
  "gopkg.in/urfave/cli.v1"
)

func main() {
  app := cli.NewApp()
	app.Version = "0.0.1"
  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}