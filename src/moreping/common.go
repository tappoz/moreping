package moreping

import (
	"io/ioutil"
	"log"
)

// Logger is the standard logger for the whole project. It can be used as
// a normal logger from the standard library via the Printf function (no other
// function has been exported).
var Logger StdLogger = log.New(ioutil.Discard, "[moreping]", log.LstdFlags)

// StdLogger is inspired by the Sarama package (a Kafka client)
// on how to use the logger from the standard library
type StdLogger interface {
	Printf(format string, v ...interface{})
}

// InPlaceAvg calculates an average value on the fly, given the previous average
// and the number of items included in the average calculation.
func InPlaceAvg(prevAvg float32, currValue float32, itemsSoFar int) float32 {
	itemsSoFarFloat := float32(itemsSoFar)
	return ((prevAvg * itemsSoFarFloat) + currValue) / (itemsSoFarFloat + float32(1))
}
