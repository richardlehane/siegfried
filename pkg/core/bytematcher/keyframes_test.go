package bytematcher

import "testing"

func TestCheckRelated(t *testing.T) {
	kfs := kfstub
	prevoff := pstub[0]
	prevkf := kfs[0]
	var ok bool
	for j, kf := range kfs[1:] {
		thisoff := pstub[j+1]
		prevoff, ok = kf.checkRelated(prevkf, thisoff, prevoff)
		if !ok {
			t.Errorf("Check related fail on kf %v", j+1)
		}
		prevkf = kf
	}
}
