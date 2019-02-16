Minimal implementation of Rabin-Karp multiple string matching algorithm. 

Example usage:

    rk, _ := rk.New([][]byte{[]byte("ab"), []byte("cd"), []byte("def")})
	for result := range rk.Index(bytes.NewBuffer([]byte("abracadabra"))) {
	  fmt.Println(result.Index, "-", result.Offset)
	}

Note: this algorithm demands a dictionary of fixed length strings. If you want variable lengths, use Aho-Corasick instead.
