export type Measure = "count" | "bytes" | "duration";

type Metric = {
  name: string,
  title?: string,
};

// Chart representing multiple metrics.
type MetricsChartConfig = {
  type: "metrics",
  measure: Measure,
  metrics: Metric[],
};

// Chart representing a single metric across multiple nodes.
type NodesChartConfig = {
  type: "nodes",
  measure: Measure,
  metric: Metric,
};

export type ChartConfig = MetricsChartConfig | NodesChartConfig;

export type DashboardConfig = {
  // Unique identifier for this dashboard, e.g. "nodes.overview"
  id: string,
  charts: ChartConfig[],
};

export function isMetricsChart(chart: ChartConfig): chart is MetricsChartConfig {
  return chart.type === "metrics";
}

export function isNodesChart(chart: ChartConfig): chart is NodesChartConfig {
  return chart.type === "nodes";
}
