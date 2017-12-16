import React from "react";
import * as api from "src/util/api";
import { connect } from "react-redux";
import { AdminUIState } from "oss/src/redux/state";
import { refreshTracesIndex } from "oss/src/redux/apiReducers";
import Link from "react-router/lib/Link";
import {CachedDataReducerState} from "oss/src/redux/cachedDataReducer";

interface TracesIndexProps {
  tracesIndexState: CachedDataReducerState<api.TracesIndexResponseMessage>;
  refreshTracesIndex: typeof refreshTracesIndex;
}

class TracesIndex extends React.Component<TracesIndexProps, {}> {

  componentDidMount() {
    this.props.refreshTracesIndex();
  }

  render() {
    return (
      <div className="section">
        <h1>Traces</h1>
        <ul>
          {!this.props.tracesIndexState.data
            ? <p>Loading...</p>
            : this.props.tracesIndexState.data.txn_idxs.map((idx) => (
                <li key={idx.toString()}>
                  <Link to={`/reports/traces/${idx.toString()}`}>{idx.toString()}</Link>
                </li>
              ))
          }
        </ul>
      </div>
    );
  }

}

function mapStateToProps(state: AdminUIState) {
  return {
    tracesIndexState: state.cachedData.tracesIndex,
  };
}

const actions = {
  refreshTracesIndex,
};

export default connect(mapStateToProps, actions)(TracesIndex);
