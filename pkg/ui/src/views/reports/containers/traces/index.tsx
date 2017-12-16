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
            : this.props.tracesIndexState.data.txn_idxs.length === 0
              ? <p>None</p>
              : this.props.tracesIndexState.data.txn_idxs.map((idx) => (
                  <li key={idx.toString()}>
                    <Link to={`/reports/traces/${idx.toString()}`}>{idx.toString()}</Link>
                  </li>
                ))
          }
        </ul>
        <h2>How To Use This</h2>
        <div>
          <pre>
            CREATE DATABASE IF NOT EXISTS traces;<br />
            CREATE TABLE IF NOT EXISTS traces.traces (<br />
            {"  "}txn_idx INT NOT NULL,<br />
            {"  "}span_idx INT NOT NULL,<br />
            {"  "}parent_span_idx INT NULL,<br />
            {"  "}message_idx INT NOT NULL,<br />
            {"  "}"timestamp" TIMESTAMP WITH TIME ZONE NOT NULL,<br />
            {"  "}duration INTERVAL NULL,<br />
            {"  "}operation STRING NULL,<br />
            {"  "}loc STRING NOT NULL,<br />
            {"  "}tag STRING NOT NULL,<br />
            {"  "}message STRING NOT NULL<br />
            );<br />
          </pre>
          <p style={{paddingTop: 15, paddingBottom: 15}}>And then</p>
          <pre>
            SET TRACING = on;<br />
            # run your query<br />
            SET TRACING = off;<br />
            INSERT INTO traces.traces (SELECT * FROM crdb_internal.session_trace);<br />
          </pre>
        </div>
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
