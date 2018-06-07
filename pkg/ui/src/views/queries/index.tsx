import _ from "lodash";
import React from "react";
import Helmet from "react-helmet";
import { connect } from "react-redux";

import Loading from "src/views/shared/components/loading";
import spinner from "assets/spinner.gif";
import { CachedDataReducerState } from "src/redux/cachedDataReducer";
import { AdminUIState } from "src/redux/state";
import { Duration } from "src/util/format";
import { FixLong } from "src/util/fixLong";
import { ColumnDescriptor, SortedTable } from "src/views/shared/components/sortedtable";
import { SortSetting } from "src/views/shared/components/sortabletable";
import { refreshQueries } from "src/redux/apiReducers";
import { QueriesResponseMessage } from "src/util/api";
import { intersperse } from "src/util/intersperse";

import * as protos from "src/js/protos";
import "./queries.styl";

type CollectedStatementStatistics$Properties = protos.cockroach.sql.CollectedStatementStatistics$Properties;
type PlanNode$Properties = protos.cockroach.sql.PlanNode$Properties;
type Attr$Properties = protos.cockroach.sql.PlanNode.Attr$Properties;

class QueriesSortedTable extends SortedTable<CollectedStatementStatistics$Properties> {}

interface QueriesPageProps {
  queries: CachedDataReducerState<QueriesResponseMessage>;
  refreshQueries: typeof refreshQueries;
}

interface QueriesPageState {
  sortSetting: SortSetting;
}

const QUERIES_COLUMNS: ColumnDescriptor<CollectedStatementStatistics$Properties>[] = [
  {
    title: "Query",
    className: "queries-table__col-query-text",
    cell: (query) => query.key.query,
    sort: (query) => query.key.query,
  },
  {
    title: "Count",
    cell: (query) => FixLong(query.stats.count).toInt(),
    sort: (query) => FixLong(query.stats.count).toInt(),
  },
  {
    title: "Avg Rows",
    cell: (query) => Math.round(query.stats.num_rows.mean),
    sort: (query) => query.stats.num_rows.mean,
  },
  {
    title: "Avg Latency",
    cell: (query) => Duration(query.stats.service_lat.mean * 1e9),
    sort: (query) => query.stats.service_lat.mean,
  },
  {
    title: "DistSQL Plan",
    cell: (query) => (
      query.stats.phys_plan_url
        ? <a href={query.stats.phys_plan_url} target="_blank">View</a>
        : null
    ),
    sort: (query) => query.stats.phys_plan_url.length, // so you can see the biggest plans
  },
  {
    title: "Plan",
    cell: (query) => (
      <PlanView plan={query.stats.most_recent_plan} />
    ),
  },
];

class QueriesPage extends React.Component<QueriesPageProps, QueriesPageState> {

  constructor(props: QueriesPageProps) {
    super(props);
    this.state = {
      sortSetting: {
        sortKey: 1,
        ascending: false,
      },
    };
  }

  changeSortSetting = (ss: SortSetting) => {
    this.setState({
      sortSetting: ss,
    });
  }

  componentWillMount() {
    this.props.refreshQueries();
  }

  componentWillReceiveProps() {
    this.props.refreshQueries();
  }

  renderQueries() {
    if (!this.props.queries.data) {
      // This should really be handled by a loader component.
      return null;
    }
    const queries = this.props.queries.data.queries;

    return (
      <div className="queries-screen">
        <span className="queries-screen__last-hour-note">
          {queries.length} query fingerprints.
          Query history is only maintained for the past hour.
        </span>

        <QueriesSortedTable
          className="queries-table"
          data={queries}
          columns={QUERIES_COLUMNS}
          sortSetting={this.state.sortSetting}
          onChangeSortSetting={this.changeSortSetting}
        />
      </div>
    );
  }

  render() {
    return (
      <section className="section" style={{ maxWidth: "none" }}>
        <Helmet>
          <title>Queries</title>
        </Helmet>

        <h1 style={{ marginBottom: 20 }}>Queries</h1>

        <Loading
          loading={_.isNil(this.props.queries.data)}
          className="loading-image loading-image__spinner"
          image={spinner}
        >
          {this.renderQueries()}
        </Loading>
      </section>
    );
  }

}

// tslint:disable-next-line:variable-name
const QueriesPageConnected = connect(
  (state: AdminUIState) => ({
    queries: state.cachedData.queries,
  }),
  {
    refreshQueries,
  },
)(QueriesPage);

export default QueriesPageConnected;

function PlanView(props: { plan: PlanNode$Properties }) {
  return (
    <ul>
      <li>
        <PlanNode node={props.plan} />
      </li>
    </ul>
  );
}

interface PlanNodeProps {
  node: PlanNode$Properties;
}

function PlanNode(props: PlanNodeProps): React.ReactElement<PlanNodeProps> {
  const node = props.node;
  node.children = node.children || [];
  node.attrs = node.attrs || [];
  const collapsedAttrs = collapseRepeatedAttrs(node.attrs);
  return (
    <div className="plan-node">
      <span className="plan-node__name">{node.name}</span>
      <span className="plan-node__attrs">
        {collapsedAttrs.map((attr) => (
          <span className="plan-node__attr" key={attr.key}>
            <span className="plan-node__attr-key">{attr.key}</span>
            <span className="plan-node__attr-eq">=</span>
            {typeof attr.value === "string"
              ? <span className="plan-node__attr-value">{attr.value}</span>
              : renderAttrValueList(attr.value as string[])}
          </span>
        ))}
      </span>
      <ul>
        {node.children.map((child, idx) => (
          <li>
            <PlanNode key={idx} node={child} />
          </li>
        ))}
      </ul>
    </div>
  );
}

function renderAttrValueList(values: string[]) {
  return (
    <span className="plan-node__attr-value-list">
      [
      {intersperse(
        values.map((value, idx) => (
          <span className="plan-node__attr-value" key={idx}>{value}</span>
        )),
        <span>, </span>,
      )}
      ]
    </span>
  );
}

interface CollapsedPlanAttr {
  // TODO(vilterp): figure out a way to make these not optional and appease the typechecker
  key?: string;
  value?: string | string[];
}

function collapseRepeatedAttrs(attrs: Attr$Properties[]): CollapsedPlanAttr[] {
  const collapsed: { [key: string]: CollapsedPlanAttr } = {};

  attrs.forEach((attr) => {
    const existingAttr = collapsed[attr.key];
    if (!existingAttr) {
      collapsed[attr.key] = attr;
      return;
    }
    if (typeof existingAttr.value === "string") {
      collapsed[attr.key] = {
        key: attr.key,
        value: [existingAttr.value, attr.value],
      };
      return;
    }
    // TODO(vilterp): type switch?
    (collapsed[attr.key].value as string[]).push(attr.value);
  });

  const collapsedAttrs = _.values(collapsed);
  return _.sortBy(collapsedAttrs, (ca) => (
    ca.key === "table" ? "table" : "z" + ca.key
  ));
}
