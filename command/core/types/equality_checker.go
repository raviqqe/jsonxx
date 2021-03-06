package types

type equalityChecker struct {
	pairs                 [][2]Type
	leftStack, rightStack []Type
}

func newEqualityChecker(s []Type) equalityChecker {
	return equalityChecker{nil, s, s}
}

func (c equalityChecker) Check(t, tt Type) bool {
	if c.isPairChecked(t, tt) {
		return true
	}

	c = c.addPair(t, tt)

	if i, ok := t.(Index); ok {
		return c.Check(c.leftStack[len(c.leftStack)-1-i.Value()], tt)
	} else if i, ok := tt.(Index); ok {
		return c.Check(t, c.rightStack[len(c.rightStack)-1-i.Value()])
	}

	switch t := t.(type) {
	case Algebraic:
		a, ok := tt.(Algebraic)

		if !ok || len(t.Constructors()) != len(a.Constructors()) {
			return false
		}

		c = c.pushTypes(t, tt)

		for i, cc := range t.Constructors() {
			es := cc.Elements()
			ees := a.Constructors()[i].Elements()

			if len(es) != len(ees) {
				return false
			}

			for i, e := range es {
				if !c.Check(e, ees[i]) {
					return false
				}
			}
		}

		return true
	case Boxed:
		b, ok := tt.(Boxed)

		if !ok {
			return false
		}

		return c.Check(t.Content(), b.Content())
	case Function:
		f, ok := tt.(Function)

		if !ok || len(t.Arguments()) != len(f.Arguments()) {
			return false
		}

		c = c.pushTypes(t, tt)

		for i, a := range t.Arguments() {
			if !c.Check(a, f.Arguments()[i]) {
				return false
			}
		}

		return c.Check(t.Result(), f.Result())
	}

	return t.equal(tt)
}

func (c equalityChecker) pushTypes(t, tt Type) equalityChecker {
	return equalityChecker{c.pairs, append(c.leftStack, t), append(c.rightStack, tt)}
}

func (c equalityChecker) addPair(t, tt Type) equalityChecker {
	return equalityChecker{append(c.pairs, [2]Type{t, tt}), c.leftStack, c.rightStack}
}

func (c equalityChecker) isPairChecked(t, tt Type) bool {
	for _, ts := range c.pairs {
		if t.equal(ts[0]) && tt.equal(ts[1]) {
			return true
		}
	}

	return false
}
