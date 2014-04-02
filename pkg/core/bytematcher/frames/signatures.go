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
