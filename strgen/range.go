package strgen

import (
	"fmt"
	"strconv"
)

type RangeIterator struct {
	items        []item
	fn           func(*RangeIterator) float64
	cycleLength  int
	currentCycle int
	tmpCycle     int
	cyclepos     int
}

func (i *RangeIterator) push(it item) {
	if it.typ == itemNumber || it.typ == itemRangeSep {
		i.items = append(i.items, it)
	}
}

func (i *RangeIterator) cycle() {
	i.tmpCycle = (i.tmpCycle + 1) % i.cyclepos
	if i.tmpCycle == 0 {
		i.currentCycle++
		if i.length() > 0 {
			i.currentCycle = i.currentCycle % i.length()
		}
	}
}

func (i *RangeIterator) get() string {
	return strconv.FormatFloat(i.fn(i), 'f', -1, 32) // FIXME
}

func (i *RangeIterator) configure() {
	chkParseError := func(i item) float64 {
		f, e := strconv.ParseFloat(i.val, 64)
		if e != nil {
			panic(fmt.Errorf("couldn't parse number: %v", i.val))
		}
		return f
	}

	switch len(i.items) {
	case 2:
		a, b := i.items[0], i.items[1]
		chkParseError(a)
		i.cycleLength = -1
		if b.typ == itemRangeSep {
			start := chkParseError(a)
			i.fn = func(i *RangeIterator) float64 {
				return start + float64(i.currentCycle)
			}
		} else {
			panic(fmt.Errorf("..x not supported as range"))
		}
	case 3:
		a := chkParseError(i.items[0])
		b := chkParseError(i.items[2])
		i.cycleLength = int(b - a)
		if i.cycleLength < 0 {
			i.cycleLength = -i.cycleLength
		}
		i.fn = func(i *RangeIterator) float64 {
			if a < b {
				return a + float64(i.currentCycle)
			}
			return a - float64(i.currentCycle)
		}
	case 5:
		a := chkParseError(i.items[0])
		b := chkParseError(i.items[2])
		c := chkParseError(i.items[4])
		i.cycleLength = int((c - a) / b)
		if i.cycleLength < 0 {
			i.cycleLength = -i.cycleLength
		}
		i.fn = func(i *RangeIterator) float64 {
			if a < c {
				return a + float64(i.currentCycle)*b
			}
			return a - float64(i.currentCycle)*b
		}
	default:
		panic(fmt.Errorf("invalid amount of range items: %d", len(i.items)))
	}
}
func (i *RangeIterator) length() int         { return i.cycleLength }
func (i *RangeIterator) finished() bool      { return i.currentCycle == 0 && i.tmpCycle == 0 }
func (i *RangeIterator) setCyclePos(pos int) { i.cyclepos = pos }
