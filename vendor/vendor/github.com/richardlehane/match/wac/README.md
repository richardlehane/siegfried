Aho-Corasick multiple string matching algorithm with choices and max offsets. 

This algorithm allows for sequences that are composed of subsequences that have max offsets (or -1 for wildcard).

Subsequences are groups of choices: any can match for the subsequence will trigger a result.

The results returned are for the matches on subsequences (NOT the full sequences). The index of those subsequences and the offset is returned.

It is up to clients to verify that the complete sequence that they are interested in has matched.

Example usage:
    
    seq := wac.Seq{
      MaxOffsets: []int64{5, -1},
      Choices: []wac.Choice{
        wac.Choice{[]byte{'b'},[]byte{'c'},[]byte{'d'}},
        wac.Choice{[]byte{'a','d'}},
        wac.Choice{[]byte{'r', 'x'}},
        []wac.Choice{[]byte{'a'}},
      }
    }
    secondSeq := wac.Seq{
      MaxOffsets: []int64{0},
      Choices: []wac.Choice{wac.Choice{[]byte{'b'}}},
    }
    w := wac.New([]wac.Seq{seq, secondSeq})
    for result := range w.Index(bytes.NewBuffer([]byte("abracadabra"))) {
  	   fmt.Println(result.Index, "-", result.Offset)
    }