package service

import (
	"io/ioutil"
	"log"
)

var Logger StdLogger = log.New(ioutil.Discard, "[DD Services]", log.LstdFlags)

// StdLogger is inspired by the Sarama package (a Kafka client)
// on how to use the logger from the standard library
type StdLogger interface {
	Printf(format string, v ...interface{})
}

func InPlaceAvg(prevAvg float32, currValue float32, itemsSoFar int) float32 {
	itemsSoFarFloat := float32(itemsSoFar)
	return ((prevAvg * itemsSoFarFloat) + currValue) / (itemsSoFarFloat + float32(1))
}
