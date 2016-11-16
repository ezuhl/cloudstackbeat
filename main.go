package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/ezuhl/cloudstackbeat/beater"
)

func main() {
	err := beat.Run("cloudstackbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
