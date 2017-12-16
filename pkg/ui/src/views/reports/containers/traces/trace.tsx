import React from "react";
import * as api from "src/util/api";
import {connect} from "react-redux";
import {AdminUIState} from "oss/src/redux/state";
import {refreshTrace} from "oss/src/redux/apiReducers";
import {CachedDataReducerState} from "oss/src/redux/cachedDataReducer";
import * as protos from "src/js/protos";
import {traceTxnIdxAttr} from "oss/src/util/constants";
import {RouterState} from "react-router";
import TraceView, { TraceViewState, initialState, Action, update } from "./TraceView";
import {cockroach} from "oss/src/js/protos";
import Span = cockroach.server.serverpb.TraceResponse.Span;
import {visitNodes} from "oss/src/views/reports/containers/traces/tree";
import Link from "react-router/lib/Link";
import SpanDetails from "oss/src/views/reports/containers/traces/SpanDetails";

interface TraceOwnProps {
  traceState: CachedDataReducerState<api.TraceResponseMessage>;
  refreshTrace: typeof refreshTrace;
}

type TraceProps = TraceOwnProps & RouterState;

interface TraceState {
  state: TraceViewState;
}

class Trace extends React.Component<TraceProps, TraceState> {

  constructor() {
    super();
    this.state = {
      state: initialState,
    };
  }

  componentDidMount() {
    this.props.refreshTrace(new protos.cockroach.server.serverpb.TraceRequest({
      txn_idx: this.props.params[traceTxnIdxAttr],
    }));
  }

  handleAction = (action: Action) => {
    this.setState({
      state: update(this.state.state, action),
    });
  }

  getHoveredSpan = (): Span => {
    const hoveredIdx = this.state.state.hoveredSpan;
    // const hoveredIdx = "0";
    if (!hoveredIdx) {
      return null;
    }
    const rootSpan = this.props.traceState.data.root_span;
    let spanWithID = null;
    visitNodes(rootSpan, (span, _) => {
      if (span.idx.toString() === hoveredIdx) {
        spanWithID = span;
      }
    });
    return spanWithID;
  }

  render() {
    return (
      <div className="section" style={{maxWidth: "none"}}>
        <h1>Trace</h1>
        <section className="section parent-link">
          <Link to="/reports/traces">Traces</Link> / Txn Idx {this.props.params[traceTxnIdxAttr]}
        </section>
        {!this.props.traceState.data || this.props.traceState.inFlight
          ? <p>Loading...</p>
          : <table>
              <tbody>
                <tr>
                  <td style={{verticalAlign: "top"}}>
                    <TraceView
                      trace={this.props.traceState.data.root_span}
                      handleAction={this.handleAction}
                      traceState={this.state.state}
                      width={1000}
                    />
                  </td>
                  <td style={{verticalAlign: "top", paddingLeft: 10}}>
                    <SpanDetails span={this.getHoveredSpan()} />
                  </td>
                </tr>
              </tbody>
            </table>
        }
      </div>
    );
  }

}

function mapStateToProps(state: AdminUIState) {
  return {
    traceState: state.cachedData.trace,
  };
}

const actions = {
  refreshTrace,
};

export default connect(mapStateToProps, actions)(Trace);
