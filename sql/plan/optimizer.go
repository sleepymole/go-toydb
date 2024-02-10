package plan

import "github.com/sleepymole/go-toydb/sql/catalog"

type optimizer interface {
	optimize(node Node) (Node, error)
}

func optimize(root Node, cat catalog.Catalog) (Node, error) {
	var err error
	optimizers := makeOptimizers(cat)
	for _, opt := range optimizers {
		root, err = opt.optimize(root)
		if err != nil {
			return nil, err
		}
	}
	return root, nil
}

func makeOptimizers(cat catalog.Catalog) []optimizer {
	return []optimizer{
		&constantFolder{},
		&filterPushdown{},
		&noopCleaner{},
		&indexLookupOptimizer{cat: cat},
		&joinTypeOptimizer{},
	}
}

type constantFolder struct{}

func (constantFolder) optimize(node Node) (Node, error) {
	panic("implement me")
}

type filterPushdown struct{}

func (filterPushdown) optimize(node Node) (Node, error) {
	panic("implement me")
}

type noopCleaner struct{}

func (noopCleaner) optimize(node Node) (Node, error) {
	panic("implement me")
}

type indexLookupOptimizer struct {
	cat catalog.Catalog
}

func (o *indexLookupOptimizer) optimize(node Node) (Node, error) {
	panic("implement me")
}

type joinTypeOptimizer struct{}

func (joinTypeOptimizer) optimize(node Node) (Node, error) {
	panic("implement me")
}
