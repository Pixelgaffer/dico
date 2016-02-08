package strgen

import "github.com/bradfitz/slice"

func GenerateStrings(optionGen string) (<-chan string, int64, error) {
	_, items := lex(optionGen)
	var iterators []Iterator
	var currIter Iterator
	emit := func() {
		currIter.configure()
		iterators = append(iterators, currIter)
		currIter = nil
	}
	for item := range items {
		switch item.typ {
		case itemRange:
			currIter = &RangeIterator{}
		case itemChoice:
			currIter = &ChoiceIterator{}
		case itemText:
			if currIter == nil {
				currIter = &TextIterator{text: item.val}
				emit()
			} else {
				currIter.push(item)
			}
		case itemIterEnd:
			emit()
		case itemEOF:
		default:
			currIter.push(item)
		}
	}

	sortable := make([]Iterator, len(iterators))
	for i := 0; i < len(iterators); i++ {
		iterators[i].configure()
		sortable[i] = iterators[i]
	}
	slice.Sort(sortable[:], func(i, j int) bool {
		return sortable[i].length() < sortable[j].length() || sortable[j].length() == -1
	})
	cyclepos := 1
	for i := 0; i < len(sortable); i++ {
		sortable[i].setCyclePos(cyclepos)
		//fmt.Printf("%+v\n", sortable[i])
		cyclepos *= sortable[i].length()
	}

	results := make(chan string)

	go func() {
		for {
			s := ""
			for i := 0; i < len(iterators); i++ {
				s += iterators[i].get()
				iterators[i].cycle()
			}
			results <- s
			if iterators[len(iterators)-1].finished() {
				close(results)
				return
			}
		}
	}()
	return results, -1, nil
}
