package fixify

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

type ModelConnectorImpl[T any] struct {
	m            T
	connectFuncs []Connecter[T]

	parentSet map[ModelConnector]struct{}
	childSet  map[ModelConnector]struct{}
}

func (mc *ModelConnectorImpl[T]) Value() T {
	return mc.m
}

type Connecter[T any] interface {
	connect(t testing.TB, childModel T, parentModel any) bool
}

func ConnectParentFunc[U, V any](f func(t testing.TB, childModel U, parentModel V)) Connecter[U] {
	return connectParentFunc[U, V](f)
}

type connectParentFunc[U, V any] func(t testing.TB, childModel U, parentModel V)

func (f connectParentFunc[U, V]) connect(t testing.TB, childModel U, parentModel any) bool {
	if v, ok := parentModel.(V); ok {
		f(t, childModel, v)
		return true
	}
	return false
}

func NewModelConnector[T any](model T, connectorFuncs ...Connecter[T]) *ModelConnectorImpl[T] {
	return &ModelConnectorImpl[T]{
		m:            model,
		connectFuncs: connectorFuncs,
	}
}

type ModelConnector interface {
	// Children() []ModelConnector
	// Descendants() []ModelConnector
	AnyValue() any
	model() any
	setParent(parent ModelConnector)
	parents() []ModelConnector
	setChild(child ModelConnector)
	children() []ModelConnector
	connectors() []func(t testing.TB, parent any) bool
}

func (mc *ModelConnectorImpl[T]) Bind(b **ModelConnectorImpl[T]) *ModelConnectorImpl[T] {
	*b = mc
	return mc
}

func (mc *ModelConnectorImpl[T]) With(connectors ...ModelConnector) *ModelConnectorImpl[T] {
	if mc.childSet == nil {
		mc.childSet = map[ModelConnector]struct{}{}
	}
	for _, c := range connectors {
		if _, ok := mc.parentSet[c]; ok {
			panic(fmt.Errorf("cyclic dependency: %T <-> %T", mc.Value(), c.model()))
		}
		mc.childSet[c] = struct{}{}
		c.setParent(mc)
	}
	return mc // メソッドチェーンで記述できるようにする
}

func (mc *ModelConnectorImpl[T]) model() any {
	return mc.m
}

func (mc *ModelConnectorImpl[T]) setParent(parent ModelConnector) {
	if mc.parentSet == nil {
		mc.parentSet = map[ModelConnector]struct{}{}
	}
	mc.parentSet[parent] = struct{}{}
}

func (mc *ModelConnectorImpl[T]) parents() []ModelConnector {
	return keys(mc.parentSet)
}

func (mc *ModelConnectorImpl[T]) setChild(child ModelConnector) {
	if mc.childSet == nil {
		mc.childSet = map[ModelConnector]struct{}{}
	}
	mc.childSet[child] = struct{}{}
}

func (mc *ModelConnectorImpl[T]) children() []ModelConnector {
	return keys(mc.childSet)
}

func (mc *ModelConnectorImpl[T]) connectors() []func(t testing.TB, parent any) bool {
	funcs := make([]func(t testing.TB, parent any) bool, 0, len(mc.connectFuncs))
	for _, f := range mc.connectFuncs {
		funcs = append(funcs, func(t testing.TB, parent any) bool {
			return f.connect(t, mc.m, parent)
		})
	}
	return funcs
}

// func (mc *ModelConnectorImpl[T]) Children() []ModelConnector {
// 	return mc.children()
// }

// func (mc *ModelConnectorImpl[T]) Descendants() []ModelConnector {
// 	return flat(mc.children()...)
// }

func (mc *ModelConnectorImpl[T]) AnyValue() any {
	return mc.m
}

type Fixture struct {
	t          testing.TB
	connectors []ModelConnector
}

func New(t testing.TB, cs ...ModelConnector) *Fixture {
	t.Helper()
	f := &Fixture{
		t: t,
	}
	allWithDuplicates := flat(cs...)
	// 順序をあえてランダムにする
	rand.Shuffle(len(allWithDuplicates), func(i, j int) {
		allWithDuplicates[i], allWithDuplicates[j] = allWithDuplicates[j], allWithDuplicates[i]
	})
	all := uniq(allWithDuplicates)
	numParents := make(map[ModelConnector]int, len(all))
	for _, c := range all {
		numParents[c] = len(c.parents())
	}
	// topological sort
	sorted := make([]ModelConnector, 0, len(all))
	for {
		allZeros := true
		for _, c := range all {
			if c != nil {
				allZeros = false
				break
			}
		}
		if allZeros {
			break
		}
		for i, c := range all {
			if c == nil {
				continue
			}
			if numParents[c] == 0 {
				sorted = append(sorted, c)
				for _, child := range c.children() {
					numParents[child] -= 1
				}
				all[i] = nil
			}
		}
	}
	f.connectors = sorted
	return f
}

// All returns all models in the fixture.
func (f *Fixture) All() []any {
	models := make([]any, 0, len(f.connectors))
	for _, c := range f.connectors {
		models = append(models, c.model())
	}
	return models
}

// Iterate applies visit in the topological order of the models.
func (f *Fixture) Iterate(setter func(model any) error) {
	f.t.Helper()
	for _, c := range f.connectors {
		if err := setter(c.model()); err != nil {
			f.t.Fatalf("failed to visit %v: %v", c.model(), err)
		}
		for _, child := range c.children() {
			ok := false
			for _, connect := range child.connectors() {
				if connect(f.t, c.model()) {
					ok = true
				}
			}
			if !ok {
				f.t.Fatalf("failed to connect: child %T to parent %T", child.model(), c.model())
			}
		}
	}
}

func uniq(cs []ModelConnector) []ModelConnector {
	seen := make(map[ModelConnector]struct{}, len(cs))
	uniq := make([]ModelConnector, 0, len(cs))
	for _, c := range cs {
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			uniq = append(uniq, c)
		}
	}
	return uniq
}

func flat(cs ...ModelConnector) []ModelConnector {
	all := make([]ModelConnector, 0, len(cs))
	for _, c := range cs {
		all = append(all, c)
		all = append(all, flat(c.children()...)...)
	}
	return all
}

func keys[U comparable, V any](m map[U]V) []U {
	ks := make([]U, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
