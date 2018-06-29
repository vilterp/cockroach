import _ from "lodash";
import React from "react";
import { createSelector } from "reselect";
import { connect } from "react-redux";
import * as Long from "long";
import Helmet from "react-helmet";

import Loading from "src/views/shared/components/loading";
import spinner from "assets/spinner.gif";
import { ToolTipWrapper } from "src/views/shared/components/toolTip";
import * as docsURL from "src/util/docs";
import { FixLong } from "src/util/fixLong";
import { TimestampToMoment } from "src/util/convert";
import { cockroach } from "src/js/protos";
import { AdminUIState } from "src/redux/state";
import {
  refreshDataDistribution,
  refreshNodes,
  refreshLiveness,
  refreshRangeLog,
} from "src/redux/apiReducers";
import { LocalityTree, selectLocalityTree } from "src/redux/localities";
import ReplicaMatrix, { SchemaObject } from "./replicaMatrix";
import { TreeNode, TreePath } from "./tree";
import { RangeLogResponseMessage } from "src/util/api";
import "./index.styl";

import RangeLogRequest = cockroach.server.serverpb.RangeLogRequest;
type DataDistributionResponse = cockroach.server.serverpb.DataDistributionResponse;
type NodeDescriptor = cockroach.roachpb.NodeDescriptor$Properties;
type ZoneConfig$Properties = cockroach.server.serverpb.DataDistributionResponse.ZoneConfig$Properties;

const ZONE_CONFIG_TEXT = (
  <span>
    Zone configurations
    (<a href={docsURL.configureReplicationZones} target="_blank">see documentation</a>)
    control how CockroachDB distributes data across nodes.
  </span>
);

interface DataDistributionProps {
  dataDistribution: DataDistributionResponse;
  localityTree: LocalityTree;
  sortedZoneConfigs: ZoneConfig$Properties[];
  rangeLog: RangeLogResponseMessage;
}

class DataDistribution extends React.Component<DataDistributionProps> {

  renderZoneConfigs() {
    return (
      <div className="zone-config-list">
        <ul>
          {this.props.sortedZoneConfigs.map((zoneConfig) => (
            <li key={zoneConfig.cli_specifier} className="zone-config">
              <h3>{zoneConfig.cli_specifier}</h3>
              <pre className="zone-config__raw-yaml">
                {zoneConfig.config_yaml}
              </pre>
            </li>
          ))}
        </ul>
      </div>
    );
  }

  renderRangeLog() {
    return (
      <div>
        <table className="range-log">
          <thead>
            <tr className="range-log__row--header">
              <th className="range-log__cell">Timestamp</th>
              <th className="range-log__cell">Event Type</th>
              <th className="range-log__cell">Event</th>
              <th className="range-log__cell">Pretty Info</th>
            </tr>
          </thead>
          <tbody>
            {this.props.rangeLog.events.map((rangeLogEvent) => (
              <tr key={rangeLogEvent.event.timestamp.seconds.toString()} className="range-log__row--body">
                <td className="range-log__cell">{TimestampToMoment(rangeLogEvent.event.timestamp).toString()}</td>
                <td className="range-log__cell">{cockroach.storage.RangeLogEventType[rangeLogEvent.event.event_type]}</td>
                <td className="range-log__cell range-log__info">{JSON.stringify(rangeLogEvent.event, null, 2)}</td>
                <td className="range-log__cell range-log__info">{JSON.stringify(rangeLogEvent.pretty_info, null, 2)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  getCellValue = (dbPath: TreePath, nodePath: TreePath): number => {
    const [dbName, tableName] = dbPath;
    const nodeID = nodePath[nodePath.length - 1];
    const databaseInfo = this.props.dataDistribution.database_info;

    const res = databaseInfo[dbName].table_info[tableName].replica_count_by_node_id[nodeID];
    if (!res) {
      return 0;
    }
    return FixLong(res).toInt();
  }

  render() {
    const nodeTree = nodeTreeFromLocalityTree("Cluster", this.props.localityTree);

    const databaseInfo = this.props.dataDistribution.database_info;
    const dbTree: TreeNode<SchemaObject> = {
      name: "Cluster",
      data: { dbName: null, tableName: null },
      children: _.map(databaseInfo, (dbInfo, dbName) => ({
        name: dbName,
        data: { dbName },
        children: _.map(dbInfo.table_info, (_tableInfo, tableName) => ({
          name: tableName,
          data: { dbName, tableName },
        })),
      })),
    };

    return (
      <div className="data-distribution">
        <div className="data-distribution__zone-config-sidebar">
          <h2>
            Zone Configs{" "}
            <div className="section-heading__tooltip">
              <ToolTipWrapper text={ZONE_CONFIG_TEXT}>
                <div className="section-heading__tooltip-hover-area">
                  <div className="section-heading__info-icon">i</div>
                </div>
              </ToolTipWrapper>
            </div>
          </h2>
          {this.renderZoneConfigs()}
        </div>
        <div>
          <ReplicaMatrix
            cols={nodeTree}
            rows={dbTree}
            getValue={this.getCellValue}
          />
          <h3>Range Log</h3>
          <div className="data-distribution__range-log">
            {this.renderRangeLog()}
          </div>
        </div>
      </div>
    );
  }
}

interface DataDistributionPageProps {
  dataDistribution: DataDistributionResponse;
  localityTree: LocalityTree;
  sortedZoneConfigs: ZoneConfig$Properties[];
  rangeLog: RangeLogResponseMessage;
  refreshDataDistribution: typeof refreshDataDistribution;
  refreshNodes: typeof refreshNodes;
  refreshLiveness: typeof refreshLiveness;
  refreshRangeLog: typeof refreshRangeLog;
}

class DataDistributionPage extends React.Component<DataDistributionPageProps> {

  componentDidMount() {
    this.refresh();
  }

  componentWillReceiveProps() {
    this.refresh();
  }

  refresh() {
    this.props.refreshDataDistribution();
    this.props.refreshNodes();
    this.props.refreshLiveness();
    this.props.refreshRangeLog(new RangeLogRequest({ range_id: Long.fromInt(0), limit: 20 }));
  }

  render() {
    return (
      <div>
        <Helmet>
          <title>Data Distribution</title>
        </Helmet>
        <section className="section">
          <h1>Data Distribution</h1>
        </section>
        <section className="section">
          <Loading
            className="loading-image loading-image__spinner-left"
            loading={!this.props.dataDistribution || !this.props.localityTree || !this.props.rangeLog}
            image={spinner}
          >
            <DataDistribution
              localityTree={this.props.localityTree}
              dataDistribution={this.props.dataDistribution}
              sortedZoneConfigs={this.props.sortedZoneConfigs}
              rangeLog={this.props.rangeLog}
            />
          </Loading>
        </section>
      </div>
    );
  }
}

const sortedZoneConfigs = createSelector(
  (state: AdminUIState) => state.cachedData.dataDistribution,
  (dataDistributionState) => {
    if (!dataDistributionState.data) {
      return null;
    }
    return _.sortBy(dataDistributionState.data.zone_configs, (zc) => zc.cli_specifier);
  },
);

// tslint:disable-next-line:variable-name
const DataDistributionPageConnected = connect(
  (state: AdminUIState) => ({
    dataDistribution: state.cachedData.dataDistribution.data,
    sortedZoneConfigs: sortedZoneConfigs(state),
    localityTree: selectLocalityTree(state),
    rangeLog: state.cachedData.rangeLog[0] ? state.cachedData.rangeLog[0].data : null,
  }),
  {
    refreshDataDistribution,
    refreshNodes,
    refreshLiveness,
    refreshRangeLog,
  },
)(DataDistributionPage);

export default DataDistributionPageConnected;

// Helpers

function nodeTreeFromLocalityTree(
  rootName: string,
  localityTree: LocalityTree,
): TreeNode<NodeDescriptor> {
  const children: TreeNode<any>[] = [];

  // Add child localities.
  _.forEach(localityTree.localities, (valuesForKey, key) => {
    _.forEach(valuesForKey, (subLocalityTree, value) => {
      children.push(nodeTreeFromLocalityTree(`${key}=${value}`, subLocalityTree));
    });
  });

  // Add child nodes.
  _.forEach(localityTree.nodes, (node) => {
    children.push({
      name: node.desc.node_id.toString(),
      data: node.desc,
    });
  });

  return {
    name: rootName,
    children: children,
  };
}
