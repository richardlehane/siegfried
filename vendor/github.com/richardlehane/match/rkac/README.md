Minimal implementation of Rabin-Karp multiple string matching algorithm. 

Example usage:

    rkac, _ := rkac.New([][]byte{[]byte("ab"), []byte("cd"), []byte("def")})
	for result := range rk.Index(bytes.NewBuffer([]byte("abracadabra"))) {
	  fmt.Println(result.Index, "-", result.Offset)
	}
