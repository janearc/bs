package main

import (
	simps "bs/simple"
	"flag"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	simple := flag.Bool("simple", false, "simple interaction")
	query := flag.String("query", "", "query to pass to the backend")
	flag.Parse()

	if *simple {
		logrus.Info("simple mode")

		// simple mode
		simp := simps.NewSimpleService(os.Getenv("SIMPLE_ROOT"))

		rsp := simp.Chat(*query)
		if rsp == "" {
			logrus.Fatalf("Whoops, looks like a little too simple")
		}

		logrus.Infof("Response: %s", rsp)
	}
}
