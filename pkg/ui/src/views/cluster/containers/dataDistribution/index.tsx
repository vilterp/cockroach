import _ from "lodash";
import React from "react";
import { connect } from "react-redux";

import Loading from "src/views/shared/components/loading";
import spinner from "assets/spinner.gif";
import { ToolTipWrapper } from "src/views/shared/components/toolTip";
import docsURL from "src/util/docs";
import { FixLong } from "oss/src/util/fixLong";
import { cockroach } from "src/js/protos";
import { AdminUIState } from "src/redux/state";
import { refreshDataDistribution, refreshNodes, refreshLiveness } from "src/redux/apiReducers";
import { LocalityTree, selectLocalityTree } from "src/redux/localities";
import Matrix from "./matrix";
import { TreeNode, TreePath } from "./tree";
import "./index.styl";

type DataDistributionResponse = cockroach.server.serverpb.DataDistributionResponse;
type NodeDescriptor = cockroach.roachpb.NodeDescriptor$Properties;

const ZONE_CONFIGS_DOCS_URL = docsURL("configure-replication-zones.html");
const ZONE_CONFIG_TEXT = (
  <span>
    Zone configurations control how CockroachDB distributes data
    across nodes. <a href={ZONE_CONFIGS_DOCS_URL} target="_blank">Docs here</a>
  </span>
);

class ReplicaMatrix extends Matrix<TableDesc, NodeDescriptor> {}

interface DataDistributionProps {
  dataDistribution: DataDistributionResponse;
  localityTree: LocalityTree;
}

class DataDistribution extends React.Component<DataDistributionProps> {

  renderZoneConfigs() {
    const zoneConfigs = this.props.dataDistribution.zone_configs;
    const sortedIDs = Object.keys(zoneConfigs);
    sortedIDs.sort();

    return (
      <div className="zone-config-list">
        <ul>
          {sortedIDs.map((zcId) => {
            const zoneConfig = zoneConfigs[zcId];
            return (
              <li key={zcId} className="zone-config">
                <h3>{zoneConfig.cli_specifier}</h3>
                <pre className="zone-config__raw-yaml">
                  {zoneConfig.config_yaml}
                </pre>
              </li>
            );
          })}
        </ul>
      </div>
    );
  }

  getCellValue = (dbPath: TreePath, nodePath: TreePath): number => {
    const [dbName, tableName] = dbPath;
    // TODO(vilterp): substring is to get rid of the "n" prefix; find a different way
    const nodeID = nodePath[nodePath.length - 1].substr(1);
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

    const dbTree: TreeNode<TableDesc> = {
      name: "Cluster",
      data: null,
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
            label="# Replicas"
            cols={nodeTree}
            rows={dbTree}
            initialCollapsedRows={[["system"]]}
            colNodeLabel={(_col, path, isPlaceholder) => (
              isPlaceholder
                ? ""
                : path.length === 0
                  ? "Cluster"
                  : path[path.length - 1]
            )}
            colLeafLabel={(node, path, isPlaceholder) => (
              isPlaceholder
                ? ""
                : node === null
                  ? path[path.length - 1]
                  : `n${node.node_id.toString()}`
            )}
            rowNodeLabel={(row: TableDesc) => (row ? `DB: ${row.dbName}` : "Cluster")}
            rowLeafLabel={(row: TableDesc) => (row.tableName)}
            getValue={this.getCellValue}
          />
        </div>
      </div>
    );
  }
}

interface DataDistributionPageProps {
  dataDistribution: DataDistributionResponse;
  localityTree: LocalityTree;
  refreshDataDistribution: typeof refreshDataDistribution;
  refreshNodes: typeof refreshNodes;
  refreshLiveness: typeof refreshLiveness;
}

class DataDistributionPage extends React.Component<DataDistributionPageProps> {

  componentDidMount() {
    this.props.refreshDataDistribution();
    this.props.refreshNodes();
    this.props.refreshLiveness();
  }

  componentWillReceiveProps() {
    this.props.refreshDataDistribution();
    this.props.refreshNodes();
    this.props.refreshLiveness();
  }

  render() {
    return (
      <div>
        <section className="section">
          <h1>Data Distribution</h1>
        </section>
        <section className="section">
          <Loading
            className="loading-image loading-image__spinner-left"
            loading={!this.props.dataDistribution || !this.props.localityTree}
            image={spinner}
          >
            <DataDistribution
              localityTree={this.props.localityTree}
              dataDistribution={this.props.dataDistribution}
            />
          </Loading>
        </section>
      </div>
    );
  }
}

// tslint:disable-next-line:variable-name
const DataDistributionPageConnected = connect(
  (state: AdminUIState) => {
    return {
      dataDistribution: state.cachedData.dataDistribution.data,
      localityTree: selectLocalityTree(state),
    };
  },
  {
    refreshDataDistribution,
    refreshNodes,
    refreshLiveness,
  },
)(DataDistributionPage);

export default DataDistributionPageConnected;

// Helpers

interface TableDesc {
  dbName: string;
  tableName?: string;
}

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
      name: `n${node.desc.node_id.toString()}`,
      data: node.desc,
    });
  });

  return {
    name: rootName,
    children: children,
  };
}
