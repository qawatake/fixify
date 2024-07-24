package fixify

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// Model is a wrapper model with ability to connect to other models.
// It implements IModel.
type Model[T any] struct {
	v              *T
	connectorFuncs []Connecter[T]

	parentSet map[imodelWithL]struct{}
	childSet  map[IModel]struct{}
}

type imodelWithL struct {
	model IModel
	label any
}

type modelWithL[T any, L comparable] struct {
	*Model[T]
	label *L
}

func (m *modelWithL[T, L]) Label() any {
	return m.label
}

type labler interface {
	Label() any
}

var _ IModel = &Model[int]{}

// IModel represents a set of models that can be connected to each other.
type IModel interface {
	// Children() []ModelConnector
	// Descendants() []ModelConnector

	model() any
	setParent(parent IModel)
	parents() []IModel
	setChild(child IModel)
	children() []IModel
	canConnect(parent any, label any) bool
	connectors() []func(t testing.TB, parent any, label any)
}

// NewModel is a constructor of Model.
func NewModel[T any](model *T, connectorFuncs ...Connecter[T]) *Model[T] {
	return &Model[T]{
		v:              model,
		connectorFuncs: connectorFuncs,
	}
}

// Connector is an interface that incorporates the connector functions of the form func(t testing.TB, childModel *U, parentModel *V).
// It is used to establish connections between different model types.
// Use [ConnectorFunc] to get one.
type Connecter[T any] interface {
	connect(t testing.TB, childModel *T, parentModel any, label any)
	canConnect(parentModel any, kind any) bool
}

// connectParentFunc[U, V] implements Connecter[U].
type connectParentFunc[U, V any] func(t testing.TB, childModel *U, parentModel *V)

var _ Connecter[int] = connectParentFunc[int, string](nil)

//nolint:unused // it is necessary to implement the interface Connecter[U].
func (f connectParentFunc[U, V]) connect(tb testing.TB, childModel *U, parentModel any, label any) {
	tb.Helper()
	if v, ok := parentModel.(*V); ok {
		f(tb, childModel, v)
	}
}

//nolint:unused // it is necessary to implement the interface Connecter[U].
func (f connectParentFunc[U, V]) canConnect(parentModel any, label any) bool {
	_, ok := parentModel.(*V)
	return ok
}

type connectParentFuncWithLabel[U, V any, L comparable] struct {
	label L
	fn    connectParentFunc[U, V]
}

//nolint:unused // it is necessary to implement the interface Connecter[U].
func (f *connectParentFuncWithLabel[U, V, L]) connect(tb testing.TB, childModel *U, parentModel any, label any) {
	tb.Helper()
	if label, ok := label.(L); ok {
		if label != f.label {
			return
		}
	}
	if v, ok := parentModel.(*V); ok {
		f.fn(tb, childModel, v)
	}
}

//nolint:unused // it is necessary to implement the interface Connecter[U].
func (f *connectParentFuncWithLabel[U, V, L]) canConnect(parentModel any, label any) bool {
	if label, ok := label.(L); ok {
		if label != f.label {
			return false
		}
	}
	_, ok := parentModel.(*V)
	return ok
}

// ConnectorFunc translates a function of the form func(t testing.TB, childModel *U, parentModel *V) into Connecter[U].
func ConnectorFunc[U, V any](f func(t testing.TB, childModel *U, parentModel *V)) Connecter[U] {
	return connectParentFunc[U, V](f)
}

// With registers children models.
func (m *Model[T]) With(children ...IModel) *Model[T] {
	for _, c := range children {
		if _, ok := m.parentSet[c]; ok {
			// cyclic dependency is not allowed because we cannot sort models in a topological order.
			panic(fmt.Errorf("cyclic dependency: %T <-> %T", m.Value(), c.model()))
		}
		if cl, ok := c.(labler); ok {
			if !c.canConnect(m.Value(), cl.Label()) {
				panic(fmt.Errorf("cannot connect: child %T -> parent %T", c.model(), m.Value()))
			}
		} else {
			if !c.canConnect(m.Value(), nil) {
				panic(fmt.Errorf("cannot connect: child %T -> parent %T", c.model(), m.Value()))
			}
		}
		m.setChild(c)
		c.setParent(m)
	}

	return m // メソッドチェーンで記述できるようにする
}

// Bind sets the pointer to the model.
// It is useful when you want to connect models to multiple parents.
func (m *Model[T]) Bind(b **Model[T]) *Model[T] {
	*b = m
	return m
}

// Value returns the underlying model.
func (m *Model[T]) Value() *T {
	return m.v
}

// model returns the underlying model.
func (m *Model[T]) model() any {
	return m.v
}

// setParent sets the parent model.
func (m *Model[T]) setParent(parent IModel) {
	if m.parentSet == nil {
		m.parentSet = map[IModel]struct{}{}
	}
	m.parentSet[parent] = struct{}{}
}

// parents returns the parent models.
func (m *Model[T]) parents() []IModel {
	return keys(m.parentSet)
}

// setChild sets the child model.
func (m *Model[T]) setChild(child IModel) {
	if m.childSet == nil {
		m.childSet = map[IModel]struct{}{}
	}
	m.childSet[child] = struct{}{}
}

// children returns the children models.
func (m *Model[T]) children() []IModel {
	return keys(m.childSet)
}

// connectors returns the connector functions.
func (m *Model[T]) connectors() []func(t testing.TB, parent any, label any) {
	funcs := make([]func(t testing.TB, parent any, label any), 0, len(m.connectorFuncs))
	for _, f := range m.connectorFuncs {
		funcs = append(funcs, func(tb testing.TB, parent any, label any) {
			tb.Helper()
			f.connect(tb, m.v, parent, label)
		})
	}
	return funcs
}

// canConnect returns true if the model can connect to the parent.
func (m *Model[T]) canConnect(parent any, label any) bool {
	for _, f := range m.connectorFuncs {
		if f.canConnect(parent, label) {
			return true
		}
	}
	return false
}

// func (m *ModelConnectorImpl[T]) Children() []ModelConnector {
// 	return m.children()
// }

// func (m *ModelConnectorImpl[T]) Descendants() []ModelConnector {
// 	return flat(m.children()...)
// }

// Fixture collects models and resolves their dependencies.
type Fixture struct {
	t          testing.TB
	connectors []IModel
}

func New(tb testing.TB, cs ...IModel) *Fixture {
	tb.Helper()
	f := &Fixture{
		t: tb,
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
					numParents[child]--
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

// Iterate applies visit and call connector functions in the topological order of the models.
func (f *Fixture) Iterate(visit func(model any) error) {
	f.t.Helper()
	for _, c := range f.connectors {
		if err := visit(c.model()); err != nil {
			f.t.Fatalf("failed to visit %v: %v", c.model(), err)
		}
		for _, child := range c.children() {
			for _, connect := range child.connectors() {
				if cl, ok := child.(labler); ok {
					connect(f.t, c.model(), cl.Label())
				} else {
					connect(f.t, c.model(), nil)
				}
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
