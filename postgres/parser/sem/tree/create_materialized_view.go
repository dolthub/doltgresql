package tree

var _ Statement = &CreateMaterializedView{}

// CreateMaterializedView represents a CREATE MATERIALIZED VIEW statement.
type CreateMaterializedView struct {
	Name        TableName
	ColumnNames NameList
	IfNotExists bool
	Using       string
	Params      StorageParams
	Tablespace  Name
	AsSource    *Select
	CheckOption ViewCheckOption
	WithNoData  bool
}

// Format implements the NodeFormatter interface.
func (node *CreateMaterializedView) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE MATERIALIZED VIEW ")
	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	ctx.FormatNode(&node.Name)

	if len(node.ColumnNames) > 0 {
		ctx.WriteByte(' ')
		ctx.WriteByte('(')
		ctx.FormatNode(&node.ColumnNames)
		ctx.WriteByte(')')
	}

	ctx.WriteString(" AS ")
	ctx.FormatNode(node.AsSource)
}

// RefreshMaterializedView represents a REFRESH MATERIALIZED VIEW statement.
type RefreshMaterializedView struct {
	Name              *UnresolvedObjectName
	Concurrently      bool
	RefreshDataOption RefreshDataOption
}

// RefreshDataOption corresponds to arguments for the REFRESH MATERIALIZED VIEW
// statement.
type RefreshDataOption int

const (
	// RefreshDataDefault refers to no option provided to the REFRESH MATERIALIZED
	// VIEW statement.
	RefreshDataDefault RefreshDataOption = iota
	// RefreshDataWithData refers to the WITH DATA option provided to the REFRESH
	// MATERIALIZED VIEW statement.
	RefreshDataWithData
	// RefreshDataClear refers to the WITH NO DATA option provided to the REFRESH
	// MATERIALIZED VIEW statement.
	RefreshDataClear
)

// Format implements the NodeFormatter interface.
func (node *RefreshMaterializedView) Format(ctx *FmtCtx) {
	ctx.WriteString("REFRESH MATERIALIZED VIEW ")
	if node.Concurrently {
		ctx.WriteString("CONCURRENTLY ")
	}
	ctx.FormatNode(node.Name)
	switch node.RefreshDataOption {
	case RefreshDataWithData:
		ctx.WriteString(" WITH DATA")
	case RefreshDataClear:
		ctx.WriteString(" WITH NO DATA")
	}
}
