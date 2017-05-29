package index

import (
	"fmt"
	"math/rand"
	"storage-fulltext/stemmer"
	"strconv"
	"testing"
)

func prepare() *Index {
	index := Create()
	stm := stemmer.NewStemmer()

	rand.Seed(42)
	for i := 0; i < 5; i++ {
		doc := Document("doc" + strconv.Itoa(i))
		words := make([]stemmer.StemResult, 0)
		for j := 0; j < 10; j++ {
			x := rand.Intn(100)
			words = append(words, stemmer.StemResult{x, j})
		}
		index.AddDocument(words, doc)
	}
	str := []string{"elo", "ziomek", "jedziemy", "ziemnik", "złomnićż", "koksem"}
	res, _ := stm.Stem(str)
	fmt.Println(res)
	doc := Document("realdoc")
	index.AddDocument(res, doc)
	return index
}

func TestIndex(t *testing.T) {
	index := prepare()
	res := index.WeightedPhraseSearch([]int{87, 24, 35, 5})
	//    fmt.Println(res)

	fmt.Println(res)

	if len(res) != 3 || res[0].Score != 6 {
		t.FailNow()
	}

	index.Deprecate("doc0")

	res = index.BowSearch([]int{87, 24, 35, 5})
	if len(res) != 2 || res[0].Score != 3 {
		t.FailNow()
	}

	index.Rebuild()

	if len(index.Mapa[5]) != 0 {
		t.FailNow()
	}

	res = index.BowSearch([]int{87, 24, 35, 5})
	if len(res) != 2 || res[0].Score != 3 {
		t.FailNow()
	}

	fmt.Println(res)

	return
}
