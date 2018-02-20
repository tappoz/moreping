package moreping_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tappoz/moreping/src/moreping"
)

var _ = Describe("Common", func() {

	Describe("In place average", func() {
		It("should calculate the mean in place with a stream of number", func() {
			sameNumSlice := []float32{10, 10, 10, 10, 10}
			// TODO check why it works only if first element is setup as current avg
			currAvg := float32(10)
			for i := 1; i < len(sameNumSlice); i++ {
				currAvg := moreping.InPlaceAvg(currAvg, sameNumSlice[i], i)
				Expect(currAvg).To(Equal(float32(10)))
			}
		})
	})
})
