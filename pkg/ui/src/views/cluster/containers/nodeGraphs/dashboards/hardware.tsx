import React from "react";

import { LineGraph } from "src/views/cluster/components/linegraph";
import { Metric, Axis, AxisUnits } from "src/views/shared/components/metricQuery";

import { GraphDashboardProps, nodeDisplayName } from "./dashboardUtils";

export default function (props: GraphDashboardProps) {
  const { nodeIDs, nodesSummary, nodeSources, tooltipSelection } = props;

  return [
    <LineGraph title="Live Node Count" tooltip="The number of live nodes in the cluster.">
      <Axis label="nodes">
        <Metric name="cr.node.liveness.livenodes" title="Live Nodes" aggregateMax />
      </Axis>
    </LineGraph>,

    <LineGraph
      title="Memory Usage"
      sources={nodeSources}
      tooltip={(
        <div>
          Memory in use {tooltipSelection}
        </div>
      )}
    >
      <Axis units={AxisUnits.Bytes} label="memory usage">
        {nodeIDs.map((nid) => (
          <Metric
            name="cr.node.sys.rss"
            title={nodeDisplayName(nodesSummary, nid)}
            sources={[nid]}
          />
        ))}
      </Axis>
    </LineGraph>,

    <LineGraph
      title="CPU Percent"
      sources={nodeSources}
      // TODO(vilterp): tooltip
    >
      {/*TODO(vilterp): wtf how do we do percent*/}
      <Axis units={AxisUnits.Count} label="CPU">
        {nodeIDs.map((nid) => (
          <Metric
            name="cr.node.sys.cpu.user.percent"
            title={nodeDisplayName(nodesSummary, nid)}
            sources={[nid]}
          />
        ))}
      </Axis>
    </LineGraph>,

    <LineGraph
      title="Disk Read Bytes"
      sources={nodeSources}
    >
      <Axis units={AxisUnits.Bytes} label="bytes">
        {nodeIDs.map((nid) => (
          <Metric
            name="cr.node.sys.disk.read.bytes.host"
            title={nodeDisplayName(nodesSummary, nid)}
            sources={[nid]}
            nonNegativeRate
          />
        ))}
      </Axis>
    </LineGraph>,

    <LineGraph
      title="Disk Write Bytes"
      sources={nodeSources}
    >
      <Axis units={AxisUnits.Bytes} label="bytes">
        {nodeIDs.map((nid) => (
          <Metric
            name="cr.node.sys.disk.write.bytes.host"
            title={nodeDisplayName(nodesSummary, nid)}
            sources={[nid]}
            nonNegativeRate
          />
        ))}
      </Axis>
    </LineGraph>,

    <LineGraph
      title="Network Bytes Received"
      sources={nodeSources}
    >
      <Axis units={AxisUnits.Bytes} label="bytes">
        {nodeIDs.map((nid) => (
          <Metric
            name="cr.node.sys.net.recv.bytes.host"
            title={nodeDisplayName(nodesSummary, nid)}
            sources={[nid]}
            nonNegativeRate
          />
        ))}
      </Axis>
    </LineGraph>,

    <LineGraph
      title="Network Bytes Sent"
      sources={nodeSources}
    >
      <Axis units={AxisUnits.Bytes} label="bytes">
        {nodeIDs.map((nid) => (
          <Metric
            name="cr.node.sys.net.send.bytes.host"
            title={nodeDisplayName(nodesSummary, nid)}
            sources={[nid]}
            nonNegativeRate
          />
        ))}
      </Axis>
    </LineGraph>,
  ];
}
