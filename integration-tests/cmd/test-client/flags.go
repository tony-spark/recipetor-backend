package main

import "flag"

var (
	flagKafkaBroker string
)

func init() {
	flag.StringVar(&flagKafkaBroker, "kafka-broker", "", "kafka broker address")
}
