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

  render() {
    return (
      <div className="section">
        <h1>Trace</h1>
        {!this.props.traceState.data || this.props.traceState.inFlight
          ? <p>Loading...</p>
          : <TraceView
              trace={this.props.traceState.data.root_span}
              handleAction={this.handleAction}
              traceState={this.state.state}
              width={1000}
            />
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
