import {cockroach} from "oss/src/js/protos";
import Span = cockroach.server.serverpb.TraceResponse.Span$Properties;

// TODO(vilterp): generic tree interface would be nice
// also share this code with TreeTable and TreeMatrix

export function visitNodes(tree: Span, fun: (node: Span, depth: number) => void) {
  function recur(node: Span, depth: number) {
    fun(node, depth);
    if (node.child_spans.length > 0) {
      node.child_spans.forEach((n) => {
        recur(n, depth + 1);
      });
    }
  }
  recur(tree, 0);
}

export function numDescendants(tree: Span) {
  let num = 0;
  visitNodes(tree, () => {
    num++;
  });
  return num;
}
