import _ from "lodash";
import React, { Component } from "react";
import classNames from "classnames";

import {
  TreeNode, TreePath, layoutTree, flatten, sumValuesUnderPaths, deepIncludes,
} from "./tree";
import "./matrix.styl";

const DOWN_ARROW = "▼";
const SIDE_ARROW = "▶";

interface MatrixState {
  collapsedRows: TreePath[];
  collapsedCols: TreePath[];
}

interface MatrixProps<R, C> {
  label: string;
  cols: TreeNode<C>;
  rows: TreeNode<R>;
  initialCollapsedRows?: TreePath[];
  getValue: (rowPath: TreePath, colPath: TreePath) => number;
  rowLeafLabel: (row: R, path?: TreePath) => string;
  rowNodeLabel: (row: R, path?: TreePath) => string;
  colLeafLabel: (col: C, path: TreePath, isPlaceholder: boolean) => string;
  colNodeLabel: (col: C, path: TreePath, isPlaceholder: boolean) => string;
}

const ROW_TREE_INDENT_PX = 18;

class Matrix<R, C> extends Component<MatrixProps<R, C>, MatrixState> {

  constructor(props: MatrixProps<R, C>) {
    super(props);
    this.state = {
      collapsedRows: props.initialCollapsedRows || [],
      collapsedCols: [],
    };
  }

  expandRow = (path: TreePath) => {
    this.setState({
      collapsedRows: this.state.collapsedRows.filter((tp) => !_.isEqual(tp, path)),
    });
  }

  collapseRow = (path: TreePath) => {
    this.setState({
      collapsedRows: [...this.state.collapsedRows, path],
    });
  }

  expandCol = (path: TreePath) => {
    this.setState({
      collapsedCols: this.state.collapsedCols.filter((tp) => !_.isEqual(tp, path)),
    });
  }

  collapseCol = (path: TreePath) => {
    this.setState({
      collapsedCols: [...this.state.collapsedCols, path],
    });
  }

  render() {
    const {
      label,
      cols,
      rows,
      getValue,
      colLeafLabel,
      colNodeLabel,
      rowLeafLabel,
      rowNodeLabel,
    } = this.props;
    const {
      collapsedRows,
      collapsedCols,
    } = this.state;

    const flattenedRows = flatten(rows, collapsedRows, true /* includeNodes */);
    const headerRows = layoutTree(cols, collapsedCols);
    const flattenedCols = flatten(cols, collapsedCols, false /* includeNodes */);

    return (
      <table className="matrix">
        <thead>
          {headerRows.map((row, idx) => (
            <tr key={idx}>
              {idx === 0
                ? <th className="matrix__metric-label">{label}</th>
                : <th />}
              {row.map((col) => {
                const colIsCollapsed = deepIncludes(collapsedCols, col.path);
                const arrow = colIsCollapsed ? SIDE_ARROW : DOWN_ARROW;
                return (
                  <th
                    key={col.path.join("/")}
                    colSpan={col.width}
                    className={classNames(
                      "matrix__column-header",
                      { "matrix__column-header--node": col.depth > 1 },
                    )}
                    onClick={() => (
                      colIsCollapsed
                        ? this.expandCol(col.path)
                        : this.collapseCol(col.path)
                    )}
                  >
                    {col.isPlaceholder
                      ? null
                      : col.depth === 1
                        ? colLeafLabel(col.data, col.path, col.isPlaceholder)
                        : `${arrow} ${colNodeLabel(col.data, col.path, col.isPlaceholder)}`}
                  </th>
                );
              })}
            </tr>
          ))}
        </thead>
        <tbody>
          {flattenedRows.map((row) => {
            const arrow = row.isCollapsed ? SIDE_ARROW : DOWN_ARROW;
            return (
              <tr
                key={JSON.stringify(row)}
                className={classNames("matrix__row", { "matrix__row--node": !row.isLeaf })}
                onClick={() => (
                  row.isCollapsed
                    ? this.expandRow(row.path)
                    : this.collapseRow(row.path)
                )}
              >
                <th
                  className={classNames(
                    "matrix__row-label",
                    { "matrix__row-label--node": !row.isLeaf },
                  )}
                  style={{ paddingLeft: row.depth * ROW_TREE_INDENT_PX + 5 }}
                >
                  {row.isLeaf
                    ? rowLeafLabel(row.data, row.path)
                    : `${arrow} ${rowNodeLabel(row.data, row.path)}`}
                </th>
                {flattenedCols.map((col) => {
                  return (
                    <td
                      key={col.path.join("/")}
                      className="matrix__cell-value"
                    >
                      {row.isLeaf || row.isCollapsed
                        ? emptyIfZero(sumValuesUnderPaths(rows, cols, row.path, col.path, getValue))
                        : null}
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
    );
  }

}

function emptyIfZero(n: number): string {
  if (n === 0) {
    return "";
  }
  return `${n}`;
}

export default Matrix;
