import React from "react";
import {cockroach} from "oss/src/js/protos";
import Span = cockroach.server.serverpb.TraceResponse.Span;
import {Duration} from "oss/src/util/format";
import {FixLong} from "oss/src/util/fixLong";
import "./index.styl";

interface SpanDetailsProps {
  span: Span;
}

function SpanDetails(props: SpanDetailsProps) {
  const clone = {...props.span};
  delete clone.child_spans;
  delete clone.log;

  if (props.span === null) {
    return null;
  }

  return (
    <div>
      <h2 style={{paddingTop: 0}}>
        {props.span.operation}
      </h2>
      <table className="span-log">
        <tbody>
          <tr>
            <td>Age</td>
            <td>{Duration(FixLong(props.span.age_ns).toNumber())}</td>
          </tr>
          <tr>
            <td>Idx</td>
            <td>{props.span.idx.toString()}</td>
          </tr>
          <tr>
            <td>Duration</td>
            <td>{Duration(props.span.duration_ns.toNumber())}</td>
          </tr>
        </tbody>
      </table>
      <h3 style={{marginTop: 10, marginBottom: 10}}>Logs</h3>
      <table className="span-log">
        <thead>
          <th>Age</th>
          <th>Message</th>
        </thead>
        <tbody>
          {props.span.log.map((logEntry) => {
            const age = FixLong(logEntry.age_ns).toNumber();
            const relAge = age - FixLong(props.span.age_ns).toNumber();
            return (
              <tr>
                <td>{Duration(relAge)}</td>
                <td>{logEntry.message}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default SpanDetails;
