import _ from "lodash";
import React, { Component } from "react";
import { numDescendants } from "./tree";
import {cockroach} from "oss/src/js/protos";
import Span = cockroach.server.serverpb.TraceResponse.Span$Properties;
import {FixLong} from "oss/src/util/fixLong";
import {Duration} from "oss/src/util/format";

const HEIGHT = 30;
const HEIGHT_PLUS_SPACE = HEIGHT + 5;

const DOWN_ARROW = "▼";
const SIDE_ARROW = "▶";

export interface Action {
  type: string;
}

const TOGGLE_COLLAPSED = "TOGGLE_COLLAPSED";
const toggleCollapsed = (spanID: string) => ({
  type: TOGGLE_COLLAPSED,
  spanID,
});

const HOVER_SPAN = "HOVER_SPAN";
const hoverSpan = (spanID: string) => ({
  type: HOVER_SPAN,
  spanID,
});

const UN_HOVER_SPAN = "UN_HOVER_SPAN";
const unHoverSpan = {
  type: UN_HOVER_SPAN,
};

export interface TraceViewState {
  hoveredSpan: string;
  collapsedSpans: string[];
}

export const initialState: TraceViewState = {
  hoveredSpan: null,
  collapsedSpans: [],
};

export function update(state: TraceViewState, action: Action) {
  switch (action.type) {
    case TOGGLE_COLLAPSED: {
      const isCollapsed = _.includes(state.collapsedSpans, action.spanID);
      return {
        ...state,
        collapsedSpans: isCollapsed
          ? state.collapsedSpans.filter((spanID) => (spanID !== action.spanID))
          : [...state.collapsedSpans, action.spanID],
      };
    }
    case HOVER_SPAN:
      return {
        ...state,
        hoveredSpan: action.spanID,
      };
    case UN_HOVER_SPAN:
      return {
        ...state,
        hoveredSpan: null,
      };
    default:
      throw new Error(`unknown type: ${action.type}`);
  }
}

function flatten(tree: Span, collapsed: string[]) {
  const output: Span[] = [];
  function recur(node: Span) {
    output.push(node);
    if (_.includes(collapsed, node.idx.toString())) {
      return;
    }
    if (node.child_spans) {
      node.child_spans.forEach((child: Span) => {
        recur(child);
      });
    }
  }
  recur(tree);
  return output;
}

function lerp(omin: number, omax: number, imin: number, imax: number) {
  return (input: number) => {
    return omin + (omax - omin) * (input - imin) / (imax - imin);
  };
}

interface TraceViewProps {
  trace: Span;
  width: number;
  traceState: TraceViewState;

  handleAction: (action: Action) => void;
}

class TraceView extends Component<TraceViewProps, {}> {

  handleAction = (action: Action) => {
    this.props.handleAction(action);
  }

  render() {
    const {
      trace,
      width,
      traceState,
    } = this.props;
    const {
      collapsedSpans,
      hoveredSpan,
    } = traceState;
    const flattened = flatten(trace, collapsedSpans);
    // TODO: don't compute this every frame
    const lastTS = _.max(flattened.map((span) => (
      FixLong(span.age_ns).toNumber() + FixLong(span.duration_ns).toNumber()
    ));
    const scale = lerp(0, width - 300, 0, lastTS);

    return (
      <svg
        width={width}
        height={flattened.length * HEIGHT_PLUS_SPACE}
        style={{ border: "1px solid black", backgroundColor: "white" }}
      >
        {flattened.map((span, idx) => {
          const isHovered = hoveredSpan === span.idx.toString();
          const isCollapsed = _.includes(collapsedSpans, span.idx.toString());
          const timeLabel = Duration(FixLong(span.duration_ns));
          const isLeaf = span.child_spans.length === 0;
          const label = isLeaf
            ? `${timeLabel} : ${span.operation}`
            : isCollapsed
              ? `${SIDE_ARROW} ${timeLabel} : ${span.operation} (${numDescendants(span)})`
              : `${DOWN_ARROW} ${timeLabel} : ${span.operation}`;
          return (
            <g
              key={span.idx.toString()}
              style={{ cursor: "pointer" }}
              onMouseOver={() => { this.handleAction(hoverSpan(span.idx.toString())); }}
              // onMouseOut={() => { this.handleAction(unHoverSpan); }}
              onClick={() => { this.handleAction(toggleCollapsed(span.idx.toString())); }}
            >
              <rect
                fill={isHovered ? "blue" : "lightblue"}
                y={idx * HEIGHT_PLUS_SPACE - 5}
                x={scale(FixLong(span.age_ns).toNumber()) + 5}
                height={HEIGHT}
                width={scale(FixLong(span.duration_ns).toNumber())}
              />
              <text
                x={scale(FixLong(span.age_ns).toNumber()) + 10}
                y={idx * HEIGHT_PLUS_SPACE + HEIGHT / 2}
                fill={isHovered ? "grey" : "black"}
              >
                {label}
              </text>
              <g>
                {span.log.map((logEntry, logIdx) => (
                  <circle
                    key={logIdx}
                    cx={scale(FixLong(logEntry.age_ns).toNumber()) + 5}
                    cy={idx * HEIGHT_PLUS_SPACE + 20}
                    r={3}
                    fill={"white"}
                    stroke={"black"}
                  />
                ))}
              </g>
            </g>
          );
        })}
      </svg>
    );
  }

}

export default TraceView;
