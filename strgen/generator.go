package strgen

import (
	"fmt"

	"github.com/bradfitz/slice"
)

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
		case itemError:
			return nil, 0, fmt.Errorf(item.val)
		default:
			currIter.push(item)
		}
	}

	sortable := make([]Iterator, len(iterators))
	for i := 0; i < len(iterators); i++ {
		err := iterators[i].configure()
		if err != nil {
			return nil, 0, err
		}
		sortable[i] = iterators[i]
	}
	slice.Sort(sortable[:], func(i, j int) bool {
		return sortable[i].length() < sortable[j].length() || sortable[j].length() == -1
	})
	isInfinite := false
	var resultAmount int64
	cyclepos := 1
	for i := 0; i < len(sortable); i++ {
		sortable[i].setCyclePos(cyclepos)
		//fmt.Printf("%+v\n", sortable[i])
		cyclepos *= sortable[i].length()
		if sortable[i].length() == -1 {
			isInfinite = true
		}
	}
	if isInfinite {
		resultAmount = -1
	} else {
		resultAmount = int64(cyclepos) // * sortable[len(sortable)-1].length())
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
	return results, resultAmount, nil
}
