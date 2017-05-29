package stemmer

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
)

type Stemmer struct {
	nextstem  int
	base      map[string]int
	stopwords map[string]bool
	sync.RWMutex
}

var oneLetter = regexp.MustCompile(`[a-zA-ZżółćęśąźńŻÓŁĆĘŚĄŹŃ]`)

func CleanString(filename string) string {
	cleanFilename := strings.Map(func(r rune) rune {
		b := oneLetter.MatchString(string(r))
		if b {
			return r
		} else {
			return -1
		}
	}, filename)
	return cleanFilename
	//return filename
}

func NewStemmer() *Stemmer {
	odms, err := os.Open("odm.txt")
	stopwords, err := os.Open("stopwords.txt")
	if err != nil {
		panic(err)
	}

	stemmer := new(Stemmer)
	stemmer.nextstem = 1
	stemmer.base = make(map[string]int)
	stemmer.stopwords = make(map[string]bool)
	stemmer.Lock()

	linereader := bufio.NewReader(odms)
	for {
		line, _, err := linereader.ReadLine()
		if err != nil {
			break
		}
		words := strings.Split(string(line), ",")
		for _, x := range words {
			if x[0] == ' ' {
				x = x[1:]
			}
			stemmer.base[removepolish(strings.ToLower(x))] = stemmer.nextstem // this is already clean
		}
		stemmer.nextstem++
	}
	logrus.Info("Created new stemmer with ", stemmer.nextstem, " words!")
	linereader = bufio.NewReader(stopwords)
	for {
		line, _, err := linereader.ReadLine()
		if err != nil {
			break
		}
		stemmer.stopwords[strings.ToLower(string(line))] = true
	}
	stemmer.Unlock()
	return stemmer
}

func (s *Stemmer) AddWord(str string) int {
	s.Lock()
	defer s.Unlock()
	finalWord := CleanString(removepolish(strings.ToLower(str)))
	s.base[finalWord] = s.nextstem
	s.nextstem++
	return s.nextstem - 1
}

type StemResult struct {
	Word, Orgpos int
}

func removepolish(s string) string {
	var buf bytes.Buffer
	for _, x := range s {
		switch x {
		case 'ą':
			buf.WriteRune('a')
		case 'ó':
			buf.WriteRune('o')
		case 'ł':
			buf.WriteRune('l')
		case 'ń':
			buf.WriteRune('n')
		case 'ć':
			buf.WriteRune('c')
		case 'ż':
			buf.WriteRune('z')
		case 'ź':
			buf.WriteRune('z')
		case 'ś':
			buf.WriteRune('s')
		case 'ę':
			buf.WriteRune('e')
		default:
			buf.WriteRune(x)
		}
	}
	return buf.String()
}

func (s *Stemmer) Stem(words []string) ([]StemResult, error) {
	s.RLock()
	defer s.RUnlock()

	hit := 0
	miss := 0

	misses := make([]string, 0)
	res := make([]StemResult, 0)
	for i, x := range words {
		x = CleanString(removepolish(strings.ToLower(x)))
		if _, ok := s.stopwords[x]; !ok {
			// if x not in stopwords
			if val, ok := s.base[x]; ok {
				res = append(res, StemResult{val, i})
				hit++
			} else {
				s.RUnlock()
				res = append(res, StemResult{s.AddWord(x), i})
				misses = append(misses, x)
				miss++
				s.RLock()
			}
		}
	}
	logrus.Info("HIT ", hit, " MISS ", miss)
	return res, nil
}
