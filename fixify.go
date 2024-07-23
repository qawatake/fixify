package fixify

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

type Model[T any] struct {
	m              *T
	connectorFuncs []Connecter[T]

	parentSet map[IModel]struct{}
	childSet  map[IModel]struct{}
}

func (mc *Model[T]) Value() *T {
	return mc.m
}

type Connecter[T any] interface {
	connect(t testing.TB, childModel *T, parentModel any)
	canConnect(parentModel any) bool
}

func ConnectorFunc[U, V any](f func(t testing.TB, childModel *U, parentModel *V)) Connecter[U] {
	return connectParentFunc[U, V](f)
}

type connectParentFunc[U, V any] func(t testing.TB, childModel *U, parentModel *V)

func (f connectParentFunc[U, V]) connect(t testing.TB, childModel *U, parentModel any) {
	if v, ok := parentModel.(*V); ok {
		f(t, childModel, v)
	}
}

func (f connectParentFunc[U, V]) canConnect(parentModel any) bool {
	_, ok := parentModel.(*V)
	return ok
}

func NewModel[T any](model *T, connectorFuncs ...Connecter[T]) *Model[T] {
	return &Model[T]{
		m:              model,
		connectorFuncs: connectorFuncs,
	}
}

type IModel interface {
	// Children() []ModelConnector
	// Descendants() []ModelConnector
	model() any
	setParent(parent IModel)
	parents() []IModel
	setChild(child IModel)
	children() []IModel
	canConnect(parent any) bool
	connectors() []func(t testing.TB, parent any)
}

func (mc *Model[T]) Bind(b **Model[T]) *Model[T] {
	*b = mc
	return mc
}

func (mc *Model[T]) With(connectors ...IModel) *Model[T] {
	for _, c := range connectors {
		if _, ok := mc.parentSet[c]; ok {
			panic(fmt.Errorf("cyclic dependency: %T <-> %T", mc.Value(), c.model()))
		}
		if !c.canConnect(mc.Value()) {
			panic(fmt.Errorf("cannot connect: child %T -> parent %T", c.model(), mc.Value()))
		}
		mc.setChild(c)
		c.setParent(mc)
	}
	return mc // メソッドチェーンで記述できるようにする
}

func (mc *Model[T]) model() any {
	return mc.m
}

func (mc *Model[T]) setParent(parent IModel) {
	if mc.parentSet == nil {
		mc.parentSet = map[IModel]struct{}{}
	}
	mc.parentSet[parent] = struct{}{}
}

func (mc *Model[T]) parents() []IModel {
	return keys(mc.parentSet)
}

func (mc *Model[T]) setChild(child IModel) {
	if mc.childSet == nil {
		mc.childSet = map[IModel]struct{}{}
	}
	mc.childSet[child] = struct{}{}
}

func (mc *Model[T]) children() []IModel {
	return keys(mc.childSet)
}

func (mc *Model[T]) connectors() []func(t testing.TB, parent any) {
	funcs := make([]func(t testing.TB, parent any), 0, len(mc.connectorFuncs))
	for _, f := range mc.connectorFuncs {
		funcs = append(funcs, func(t testing.TB, parent any) {
			f.connect(t, mc.m, parent)
		})
	}
	return funcs
}

func (mc *Model[T]) canConnect(parent any) bool {
	for _, f := range mc.connectorFuncs {
		if f.canConnect(parent) {
			return true
		}
	}
	return false
}

// func (mc *ModelConnectorImpl[T]) Children() []ModelConnector {
// 	return mc.children()
// }

// func (mc *ModelConnectorImpl[T]) Descendants() []ModelConnector {
// 	return flat(mc.children()...)
// }

type Fixture struct {
	t          testing.TB
	connectors []IModel
}

func New(t testing.TB, cs ...IModel) *Fixture {
	t.Helper()
	f := &Fixture{
		t: t,
	}
	allWithDuplicates := flat(cs)
	// 順序をあえてランダムにする
	rand.Shuffle(len(allWithDuplicates), func(i, j int) {
		allWithDuplicates[i], allWithDuplicates[j] = allWithDuplicates[j], allWithDuplicates[i]
	})
	all := uniq(allWithDuplicates)
	numParents := make(map[IModel]int, len(all))
	for _, c := range all {
		numParents[c] = len(c.parents())
	}
	// topological sort
	sorted := make([]IModel, 0, len(all))
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
			for _, connect := range child.connectors() {
				connect(f.t, c.model())
			}
		}
	}
}

func uniq(cs []IModel) []IModel {
	seen := make(map[IModel]struct{}, len(cs))
	uniq := make([]IModel, 0, len(cs))
	for _, c := range cs {
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			uniq = append(uniq, c)
		}
	}
	return uniq
}

func flat(cs []IModel) []IModel {
	all := make([]IModel, 0, len(cs))
	for _, c := range cs {
		all = append(all, c)
		all = append(all, flat(c.children())...)
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
