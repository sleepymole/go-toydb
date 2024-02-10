package catalog

type Table struct {
	Name    string
	Columns []*Column
}

func (t *Table) String() string {
	panic("implement me")
}

func NewTable(name string, columns []*Column) *Table {
	return &Table{
		Name:    name,
		Columns: columns,
	}
}

func (t *Table) GetColumn(name string) (*Column, error) {
	panic("implement me")
}

func (t *Table) GetColumnIndex(name string) (int, error) {
	panic("implement me")
}
