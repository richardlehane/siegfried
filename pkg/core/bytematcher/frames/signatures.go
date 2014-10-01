package frames

// Signatures are just slices of frames
type Signature []Frame

func (s Signature) String() string {
	var str string
	for i, v := range s {
		if i > 0 {
			str += " | "
		}
		str += v.String()
	}
	return "(" + str + ")"
}

func (s Signature) Equals(s1 Signature) bool {
	if len(s) != len(s1) {
		return false
	}
	for i, v := range s {
		if !v.Equals(s1[i]) {
			return false
		}
	}
	return true
}
