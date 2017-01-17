package table

import (
	"fmt"

	"github.com/freneticmonkey/migrate/go/util"
)

// DiffNode Used to define a node in the dependency DAG
type DiffNode struct {
	Diff      Diff
	Columns   []string
	Traversed bool
	LoopCheck bool
	DependsOn []*DiffNode
	Blocking  []*DiffNode
}

// NewDiffNode Helper function for initialising a DiffNode
func NewDiffNode(diff Diff) DiffNode {
	df := DiffNode{
		Diff: diff,
	}
	return df
}

// Name Builds a Diff DAG compatible name for the DiffNode
func (df DiffNode) Name() string {
	return getDiffName(df.Diff)
}

// IsBlocking Does this node have a blocking relationship with another node.
func (df DiffNode) IsBlocking() bool {
	return len(df.Blocking) > 0
}

// AddDependency Add dependency on another node.
func (df *DiffNode) AddDependency(dep *DiffNode) {
	df.DependsOn = append(df.DependsOn, dep)
	dep.AddBlock(df)
}

// AddBlock Mark this node as blocking another node.
func (df *DiffNode) AddBlock(block *DiffNode) {
	df.Blocking = append(df.Blocking, block)
}

// AddColumn Assign a relationship to a Column to this node.
func (df *DiffNode) AddColumn(column string) {
	df.Columns = append(df.Columns, column)
}

// UsesColumn Helper function. Does have a relationship with the Column indicated by the parameter
func (df DiffNode) UsesColumn(column string) (usesColumn bool) {
	for _, col := range df.Columns {
		if col == column {
			return true
		}
	}
	return false
}

// visit DAG walking function
func visit(node *DiffNode, sortedDiffs *Differences) (err error) {

	if node.LoopCheck {
		return fmt.Errorf("Diff Sort.  Unable to resolve difference sort order due to interdependent operations")
	}

	if !node.Traversed {
		node.LoopCheck = true

		for _, depNode := range node.DependsOn {

			err = visit(depNode, sortedDiffs)
			if err != nil {
				return err
			}
		}
		node.Traversed = true
		node.LoopCheck = false

		// Add Diff to end of diffs
		sortedDiffs.Slice = append(sortedDiffs.Slice, node.Diff)
	}

	return nil
}

// getDiffName Helper function. Generate a diff name using operation type to avoid conflicting diffs
func getDiffName(diff Diff) string {
	name := diff.Property
	switch {
	case diff.Op == Add:
		return name + "_add"
	case diff.Op == Del:
		return name + "_del"
	case diff.Op == Mod:
		return name + "_mod"
	}
	return name
}

// orderDiffs Post Process sort the diff operations by building and traversing
// a Directed Acyclic Graph. This is intended to prevent issues such as columns
// being dropped before an associated index is removed / updated, or indexes
// being created before columns exist
func orderDiffs(diffs Differences, forward bool) (orderedDiffs Differences, err error) {

	// If there's only a single item, early out
	if len(diffs.Slice) < 2 {
		util.LogAttentionf("NOT Ordering Table Diffs: %s", diffs.Slice[0].Table)
		return diffs, nil
	}

	// This algorithm assumes that indexes are modified in parallel to columns
	// i.e. if a column is removed, there will be a matching diff which removes
	//      the column from the index

	// Check for indexes

	hasIndexes := false

	indexDiffs := make(map[string]*DiffNode)
	colDiffs := make(map[string]*DiffNode)

	for _, diff := range diffs.Slice {

		if diff.Field == "PrimaryIndex" || diff.Field == "SecondaryIndexes" {
			hasIndexes = true

			diffNode := NewDiffNode(diff)

			// Indexes are only Mod if the AutoInc property is being changed.
			if diff.Op == Mod {

				// Build the list of Columns used by the index
				dp, ok := diff.Value.(DiffPair)
				if !ok {
					util.LogErrorf("Problem extracting DiffPair from Index Diff")
				}

				fromInd := dp.From.(Index)
				toInd := dp.To.(Index)
				for _, col := range fromInd.Columns {
					diffNode.AddColumn(col.Name)
					util.LogErrorf("From Index: %s using column: %s", fromInd.Name, col.Name)
				}

				for _, col := range toInd.Columns {
					diffNode.AddColumn(col.Name)
					util.LogErrorf("To Index: %s using column: %s", toInd.Name, col.Name)
				}

			} else {

				index, ok := diff.Value.(Index)

				if !ok {
					util.LogErrorf("Problem extracting Index from Index Diff")
				}

				for _, col := range index.Columns {
					util.LogErrorf("Adding Index: %s using column: %s", index.Name, col.Name)
					diffNode.AddColumn(col.Name)
				}
			}
			indexDiffs[getDiffName(diff)] = &diffNode

		} else if diff.Field == "Columns" {
			diffNode := NewDiffNode(diff)
			diffNode.AddColumn(diff.Property)
			colDiffs[getDiffName(diff)] = &diffNode
		}
	}

	// if indexes and there are also columns diffs occurring, build a list of columns
	if hasIndexes && len(colDiffs) > 0 {

		indexesIndependent := true

		// Post Process Add/Del Index DiffNodes.  Add is dependent on Del but can only be processed after all DiffNodes have been created.

		// Check for any Add Index Diffs
		for i, ind := range indexDiffs {
			if ind.Diff.Op == Add {

				delName := ind.Diff.Property + "_del"
				// Get a matching Del DiffNode for the index
				idf, ok := indexDiffs[delName]
				// If found
				if ok {
					indexDiffs[i].AddDependency(idf)
					indexesIndependent = false
				}
			}
		}

		// check other operations for diffs on columns used indexes
		for _, diffNode := range colDiffs {
			colName := diffNode.Diff.Property

			// Check if the column is used by any index operations
			for _, ind := range indexDiffs {
				if ind.UsesColumn(colName) {

					// If it does, then the indexes are not independent
					indexesIndependent = false
					// halt check immediately
					break
				}
			}
		}

		// if the indexes are not independent
		if !indexesIndependent {

			// Rules of dependency
			//
			// ADD Col -> None
			// DEL/MOD Col -> DEL/MOD Index
			//
			// ADD/MOD Ind -> ADD/MOD Col
			// DEL Ind -> None

			// otherwise
			nodeRoot := DiffNode{
				Diff: Diff{
					Property: "ROOT",
				},
			}

			for _, diff := range diffs.Slice {

				if (diff.Op == Add || diff.Op == Mod) && (diff.Field == "PrimaryIndex" || diff.Field == "SecondaryIndexes") {
					// Get the current index DiffNode
					idf, ok := indexDiffs[getDiffName(diff)]

					if ok {
						// search for column operations on any of the columns in the index
						for _, icol := range idf.Columns {

							// if found an Add
							cd, found := colDiffs[icol+"_add"]
							if found {
								// there's a dependency on the column
								idf.AddDependency(cd)
							}

							// if found a Del
							cd, found = colDiffs[icol+"_del"]
							if found {
								// there's a dependency on the column
								idf.AddDependency(cd)
							}
						}
					}
				}

				if (diff.Op == Del || diff.Op == Mod) && (diff.Field == "Columns") {

					// search for column operations
					// get the current column DiffNode
					cd := colDiffs[getDiffName(diff)]

					// Check each index
					for _, idxDiff := range indexDiffs {

						// if the index is using this column
						if idxDiff.UsesColumn(diff.Property) {
							// Column is dependent on the index
							cd.AddDependency(idxDiff)
						}
					}
				}
			}

			// Now check each diffs array and add any nodes that aren't blocking or have dependencies to the root.
			for i, idxDiff := range indexDiffs {
				if !idxDiff.IsBlocking() {
					nodeRoot.AddDependency(indexDiffs[i])
				}
			}

			for i, colDiff := range colDiffs {
				if !colDiff.IsBlocking() {
					nodeRoot.AddDependency(colDiffs[i])
				}
			}

			// Traverse the dependency graph

			var sortedDiffs Differences

			for _, node := range nodeRoot.DependsOn {
				visit(node, &sortedDiffs)
			}

			orderedDiffs = sortedDiffs

		} else {
			util.LogWarnf("Operations are independent.  Sorting is not required.")
			orderedDiffs = diffs
		}
	}

	return orderedDiffs, err
}
