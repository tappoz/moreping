package moreping_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/tappoz/moreping/src/moreping"
)

func TestInPlaceAvg(t *testing.T) {
	g := NewGomegaWithT(t)

	sameNumSlice := []float32{10, 10, 10, 10, 10}
	// TODO check why it works only if first element is setup as current avg
	currAvg := float32(10)
	for i := 1; i < len(sameNumSlice); i++ {
		currAvg := moreping.InPlaceAvg(currAvg, sameNumSlice[i], i)
		g.Expect(currAvg).To(Equal(float32(10)))
	}
}
