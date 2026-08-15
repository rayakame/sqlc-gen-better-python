package transform

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/sqllex"
)

// rewriteSQLiteSQL strips the index from sqlc's numbered SQLite placeholders,
// leaving every bind slot as a bare "?".
//
// SQLite binds "?N" to slot N and a bare "?" to one past the highest slot it
// has seen, and sqlc numbers its placeholders assuming each sqlc.slice marker
// occupies exactly one slot. The marker expands at call time into one slot per
// element - or none for an empty sequence - so every "?N" after a marker names
// the wrong slot, and the arguments land on the wrong columns or the statement
// is rejected outright. Bare placeholders have no such coupling: SQLite
// numbers them left to right, which is exactly the order the drivers pass
// arguments in, whatever a marker expanded to.
func rewriteSQLiteSQL(sql string) string {
	var out strings.Builder
	out.Grow(len(sql))
	for _, token := range sqllex.Scan(sql, sqllex.SQLite) {
		if token.Kind == sqllex.KindPlaceholder && token.Number != 0 {
			out.WriteString(sqllex.SQLite.Placeholder())

			continue
		}
		out.WriteString(sql[token.Start:token.End])
	}

	return out.String()
}

// needsSlotRewrite reports whether the query's numbering has to be undone. It
// takes both halves of the problem: sqlc numbers placeholders as soon as a
// query uses a named argument, and only a sqlc.slice marker - the one bind
// slot whose width is unknown until call time - can desynchronize them.
// Without a marker sqlc's indexes hold, and the SQL is left alone.
func needsSlotRewrite(slots []sqllex.Slot) bool {
	numbered, sliced := false, false
	for _, slot := range slots {
		switch {
		case slot.Name != "":
			sliced = true
		case slot.Number != 0:
			numbered = true
		}
	}

	return numbered && sliced
}

// reorderBindOrder puts the query's arguments in bind-slot order, reporting
// whether it could. Callers must leave the SQL numbered when it could not:
// sqlc's order and sqlc's numbering agree with each other, so a query this
// rewrite does not understand stays exactly as it was.
func reorderBindOrder(query *model.Query, slots []sqllex.Slot) bool {
	if bundled := bundledTable(query.Params); bundled != nil {
		ordered, ok := orderColumnsBySlot(bundled.Columns, slots)
		if ok {
			bundled.Columns = ordered
		}

		return ok
	}
	ordered, ok := orderParamsBySlot(query.Params, slots)
	if ok {
		query.Params = ordered
	}

	return ok
}

// bundledTable returns the Params class the query's parameters were bundled
// into, or nil when the query takes its parameters one by one.
func bundledTable(params []model.QueryValue) *model.Table {
	for _, param := range params {
		if param.EmitTable {
			return param.Table
		}
	}

	return nil
}

// orderParamsBySlot rebuilds a plain parameter list in bind-slot order.
func orderParamsBySlot(params []model.QueryValue, slots []sqllex.Slot) ([]model.QueryValue, bool) {
	return orderBySlot(params, slots,
		func(param model.QueryValue) (string, int32) { return param.Type.SqlcSliceName, param.Number },
		func(param *model.QueryValue) { param.Repeated = true },
	)
}

// orderColumnsBySlot rebuilds the fields of a bundled Params class in bind-slot
// order. The drivers expand the class positionally, field by field, so the
// field order is the bind order.
func orderColumnsBySlot(columns []model.Column, slots []sqllex.Slot) ([]model.Column, bool) {
	return orderBySlot(columns, slots,
		func(column model.Column) (string, int32) { return column.Type.SqlcSliceName, column.Number },
		func(column *model.Column) { column.Repeated = true },
	)
}

// orderBySlot rebuilds a bind list in slot order, which is what positional
// binding needs and what sqlc's own order is not: with a numbered query sqlc
// sorts parameters by index, and an explicit index can sit anywhere in the
// text.
//
// An item bound at several slots is emitted once per slot, the later ones
// flagged Repeated so they keep their binding slot without repeating in the
// signature or the Params class. A sqlc.slice is the exception: its marker
// expands to as many slots as the sequence is long, and the drivers already
// replay its argument per marker occurrence, so it is emitted once.
//
// It reports false when the scan and sqlc's parameters do not line up exactly.
// Guessing there would bind an argument that does not exist.
func orderBySlot[T any](items []T, slots []sqllex.Slot, keyOf func(T) (string, int32), markRepeated func(*T)) ([]T, bool) {
	byNumber := make(map[int]int, len(items))
	bySlice := make(map[string]int, len(items))
	for i, item := range items {
		sliceName, number := keyOf(item)
		if sliceName != "" {
			bySlice[sliceName] = i
		}
		byNumber[int(number)] = i
	}

	ordered := make([]T, 0, len(slots))
	seen := make(map[int]struct{}, len(items))
	for _, slot := range slots {
		var idx int
		var found bool
		switch {
		case slot.Name != "":
			idx, found = bySlice[slot.Name]
			if _, repeat := seen[idx]; found && repeat {
				// The marker occurs again; the driver replays the argument.
				continue
			}
		case slot.Number != 0:
			idx, found = byNumber[slot.Number]
		default:
			// sqlc numbers every placeholder of a query that has a named
			// argument, so the only unnumbered slot is a marker's. Anything
			// else is a shape whose bind order can only be guessed at.
			return nil, false
		}
		if !found {
			// sqlc reported nothing for this slot (it aliases a slice name, or
			// dropped the parameter).
			return nil, false
		}
		if sliceName, _ := keyOf(items[idx]); slot.Name == "" && sliceName != "" {
			// A plain slot resolving to a slice means sqlc aliased an argument
			// and a sqlc.slice sharing one name into a single parameter,
			// leaving this slot unbindable. Not ours to repair.
			return nil, false
		}
		item := items[idx]
		if _, repeat := seen[idx]; repeat {
			markRepeated(&item)
		}
		seen[idx] = struct{}{}
		ordered = append(ordered, item)
	}
	if len(seen) != len(items) {
		// Something binds nowhere the scan could see.
		return nil, false
	}

	return ordered, true
}
