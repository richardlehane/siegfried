Minimal implementation of Aho-Corasick multiple string matching algorithm. 

Example usage:

    ac := ac.New([][]byte{[]byte("ab"), []byte("c"), []byte("def")})
	for result := range ac.Index(bytes.NewBuffer([]byte("abracadabra"))) {
	  fmt.Println(result.Index, "-", result.Offset)
	}

This implementation is tuned for fast matching speed. Building the Aho-Corasick tree is relatively slow and memory intensive and it only returns the index (within the byte slices that made the tree) and offset of matches. For a more fully featured and balanced implementation, use [http://godoc.org/code.google.com/p/ahocorasick](http://godoc.org/code.google.com/p/ahocorasick).
