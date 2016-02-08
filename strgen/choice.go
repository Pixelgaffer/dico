package strgen

type ChoiceIterator struct {
	choices      []string
	currentCycle int
	tmpCycle     int
	cyclepos     int
}

func (i *ChoiceIterator) push(it item) {
	if it.typ == itemText {
		i.choices = append(i.choices, it.val)
	}
}
func (i *ChoiceIterator) cycle() {
	i.tmpCycle = (i.tmpCycle + 1) % i.cyclepos
	if i.tmpCycle == 0 {
		i.currentCycle = (i.currentCycle + 1) % i.length()
	}
}
func (i *ChoiceIterator) get() string          { return i.choices[i.currentCycle] }
func (i *ChoiceIterator) configure() (e error) { return }
func (i *ChoiceIterator) length() int          { return len(i.choices) }
func (i *ChoiceIterator) finished() bool       { return i.currentCycle == 0 && i.tmpCycle == 0 }
func (i *ChoiceIterator) setCyclePos(pos int)  { i.cyclepos = pos }
