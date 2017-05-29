package index

import (
	"sort"
	"storage-fulltext/common"
	"storage-fulltext/stemmer"
	"sync"

	"github.com/Sirupsen/logrus"
	"time"
)

const TOP_RESULTS_RETURNED = 100

type Document string

type Occurence struct {
	Doc       Document
	Pos, Prev int
}

type IndexMeta struct {
	WordsUsed, ListedOccurences int
	LastRefresh                 bool // oczywiście to nie bool ale jakaś data
}

type SearchResult struct {
	Doc                   Document
	PosBeg, PosEnd, Score int
}

type ResSlice []SearchResult

func (rs ResSlice) Len() int {
	return len(rs)
}

func (rs ResSlice) Less(i, j int) bool {
	return rs[i].Score < rs[j].Score
}

func (rs ResSlice) Swap(i, j int) {
	x := rs[i]
	rs[i] = rs[j]
	rs[j] = x
}

type Index struct {
	Mapa       map[int][]Occurence
	Deprecated map[Document]bool
	Info       IndexMeta
	sync.RWMutex
}

func Create() *Index {
	// FIXME niekonsekwencja NewIndex?
	index := new(Index)
	index.Mapa = make(map[int][]Occurence)
	index.Deprecated = make(map[Document]bool)
	return index
}

func (idx *Index) AddDocument(words []stemmer.StemResult, doc Document) {
	idx.Lock()
	defer idx.Unlock()
	prev := -1
	for _, x := range words {
		idx.Mapa[x.Word] = append(idx.Mapa[x.Word], Occurence{doc, x.Orgpos, prev})
		prev = x.Orgpos
	}
}

/*
func (idx *Index) PhraseSearch(words []int) ResSlice {
    res := make(map[Document]int)
    for _, word := range words {
        for _ , doc := range idx.Mapa[word] {
            res[doc]++
        }
    }
    sl := make(ResSlice, 0)
    for k, v := range res {
        sl = append(sl, SearchResult{Doc:k, Score:v})
    }
    sort.Sort(sl)
    cnt := common.MinInt(len(sl), TOP_RESULTS_RETURNED)
    sl = sl[:cnt]
    return sl
}
*/

type WeightedRes struct {
	Beg, Pos, Score, Boost int
}

func (idx *Index) WeightedPhraseSearch(words []int) ResSlice {
	res := make(map[Document]WeightedRes)

	idx.RLock()
	for _, word := range words {
		for _, hit := range idx.Mapa[word] {
			// iteruj po dokumentach zawierających słowo
			if val, ok := res[hit.Doc]; ok {
				// dokument jest w rezultach
				val.Score++ // zwiększ score dokmentu w rezultacie
				if val.Pos == hit.Prev {
					// jeśli ostatnie znalezione słowo to poprzednik słowa w dokumencie
					val.Boost++ // zwiększ jego ważność
				}
				val.Pos = hit.Pos // zupdate'uj pozycje ostatniego znalezionego słowa
				res[hit.Doc] = val // nadpisz element, bo po value
			} else {
				val := WeightedRes{
					Score:1,
					Boost:1,
					Beg:hit.Pos, // początek znalezionej frazy
					Pos:hit.Pos, // pozycja ostatniego znalezionego słowa
				}
				res[hit.Doc] = val
			}
		}
	}
	sl := make(ResSlice, 0)
	for k, v := range res {
		if idx.Deprecated[k] == true {
			// co kto lubi, jak kto woli
			continue
		}
		sl = append(sl, SearchResult{Doc:k, Score:v.Score * v.Boost, PosBeg: v.Beg, PosEnd: v.Pos})
	}
	idx.RUnlock()
	sort.Sort(sort.Reverse(sl)) // FIXME, confer deklaracje dla ResSlice
	cnt := common.MinInt(len(sl), TOP_RESULTS_RETURNED)
	sl = sl[:cnt]
	return sl
}

type Bow struct {
	Cnt int
	Doc Document
}

func (idx *Index) BowSearch(words []int) ResSlice {
	res := make(map[Bow]int)

	idx.RLock()
	for _, word := range words {
		for _, hit := range idx.Mapa[word] {
			bow := Bow{Cnt: hit.Pos / common.BOW_SIZE, Doc: hit.Doc}
			if val, ok := res[bow]; ok {
				val++
				res[bow] = val
			} else {
				res[bow] = 1
			}
		}
	}
	sl := make(ResSlice, 0)
	for k, v := range res {
		if idx.Deprecated[k.Doc] == true {
			continue
		}
		sl = append(sl, SearchResult{Doc:k.Doc, Score:v, PosBeg:k.Cnt * common.BOW_SIZE, PosEnd:(k.Cnt + 1) * common.BOW_SIZE})
	}
	idx.RUnlock()
	sort.Sort(sort.Reverse(sl))
	cnt := common.MinInt(len(sl), TOP_RESULTS_RETURNED)
	sl = sl[:cnt]
	return sl
}

//  Deprecation - simple "removal" method. Document exist in index, but deprecated results are not
//  pushed into results. They will be removed at next index rebuild.
func (idx *Index) Deprecate(doc Document) {
	logrus.Debug("deprecating", doc)
	idx.Lock()
	defer idx.Unlock()
	idx.Deprecated[doc] = true
	return
}

func (idx *Index) Rebuild() {
	idx.Lock()
	defer idx.Unlock()
	for k, vsl := range idx.Mapa {
		nslice := make([]Occurence, 0)
		for _, occ := range vsl {
			if val, ok := idx.Deprecated[occ.Doc]; !ok || !val {
				nslice = append(nslice, occ)
			}
		}
		idx.Mapa[k] = nslice
	}
	idx.Deprecated = make(map[Document]bool)
}

func (idx *Index) PrintMeta() {
	idx.Lock()
	defer idx.Unlock()

	start := time.Now()

	idx.Info.WordsUsed = 0
	idx.Info.ListedOccurences = 0
	keysVisited := 0

	for _, val := range idx.Mapa {
		keysVisited ++
		if l := len(val); l > 0 {
			idx.Info.WordsUsed ++
			idx.Info.ListedOccurences += l
		}
	}

	elapsed := time.Since(start)

	logrus.Printf("%d words used, %d listed occurences. %f seconds wasted.\n", idx.Info.WordsUsed,
		idx.Info.ListedOccurences, elapsed.Seconds())
}
