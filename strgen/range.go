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

func (i *RangeIterator) configure() error {
	chkParseError := func(i item) (float64, error) {
		f, e := strconv.ParseFloat(i.val, 64)
		if e != nil {
			return 0, fmt.Errorf("couldn't parse number: %v", i.val)
		}
		return f, nil
	}

	switch len(i.items) {
	case 2:
		a, b := i.items[0], i.items[1]
		chkParseError(a)
		i.cycleLength = -1
		if b.typ == itemRangeSep {
			start, err := chkParseError(a)
			if err != nil {
				return err
			}
			i.fn = func(i *RangeIterator) float64 {
				return start + float64(i.currentCycle)
			}
		} else {
			return fmt.Errorf("..x not supported as range")
		}
	case 3:
		a, err := chkParseError(i.items[0])
		if err != nil {
			return err
		}
		b, err := chkParseError(i.items[2])
		if err != nil {
			return err
		}
		if a > b {
			i.cycleLength = int(a - b + 1)
		} else {
			i.cycleLength = int(b - a + 1)
		}
		i.fn = func(i *RangeIterator) float64 {
			if a < b {
				return a + float64(i.currentCycle)
			}
			return a - float64(i.currentCycle)
		}
	case 5:
		a, err := chkParseError(i.items[0])
		if err != nil {
			return err
		}
		b, err := chkParseError(i.items[2])
		if err != nil {
			return err
		}
		c, err := chkParseError(i.items[4])
		if err != nil {
			return err
		}
		if a > c {
			if b >= 0 {
				return fmt.Errorf("sequence doesnt reach its end")
			}
			i.cycleLength = int((a-c)/(-b) + 1)
		} else {
			if b <= 0 {
				return fmt.Errorf("sequence doesnt reach its end")
			}
			i.cycleLength = int((c-a)/b + 1)
		}
		i.fn = func(i *RangeIterator) float64 {
			if a < c {
				return a + float64(i.currentCycle)*b
			}
			return a - float64(i.currentCycle)*b
		}
	default:
		return fmt.Errorf("invalid amount of range items: %d", len(i.items))
	}
	return nil
}
func (i *RangeIterator) length() int         { return i.cycleLength }
func (i *RangeIterator) finished() bool      { return i.currentCycle == 0 && i.tmpCycle == 0 }
func (i *RangeIterator) setCyclePos(pos int) { i.cyclepos = pos }
