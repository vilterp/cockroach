import d3 from "d3";
import React from "react";
import moment from "moment";
import * as nvd3 from "nvd3";
import { createSelector } from "reselect";

import * as protos from  "src/js/protos";
import { HoverState, hoverOn, hoverOff } from "src/redux/hover";
import { findChildrenOfType } from "src/util/find";
import {
  ConfigureLineChart, InitLineChart, CHART_MARGINS,
} from "src/views/cluster/util/graphs";
import {
  Metric, MetricProps, Axis, AxisProps,
} from "src/views/shared/components/metricQuery";
import { MetricsDataComponentProps } from "src/views/shared/components/metricQuery";
import Visualization from "src/views/cluster/components/visualization";
import { NanoToMilli } from "src/util/convert";

type TSDatapoint = protos.cockroach.ts.tspb.TimeSeriesDatapoint$Properties;

interface LineGraphProps extends MetricsDataComponentProps {
  title?: string;
  subtitle?: string;
  legend?: boolean;
  xAxis?: boolean;
  tooltip?: React.ReactNode;
  hoverOn?: typeof hoverOn;
  hoverOff?: typeof hoverOff;
  hoverState?: HoverState;
  chartKey?: string;
}

// Find which data point is closest to a specific time.
function bisectSeries(datapoints: TSDatapoint[], time: number) {
  if (!datapoints || datapoints.length === 0) {
    return;
  }

  const series = datapoints.map((d) => NanoToMilli(d.timestamp_nanos.toNumber()));

  const right = d3.bisectRight(series, time);
  const left = right - 1;

  let index = 0;

  if (right >= series.length) {
    // We're hovering over the rightmost point.
    index = left;
  } else if (left < 0) {
    // We're hovering over the leftmost point.
    index = right;
  } else {
    // The general case: we're hovering somewhere over the middle.
    const leftDistance = time - series[left];
    const rightDistance = series[right] - time;

    index = leftDistance < rightDistance ? left : right;
  }

  return moment(new Date(series[index]));
}

/**
 * LineGraph displays queried metrics in a line graph. It currently only
 * supports a single Y-axis, but multiple metrics can be graphed on the same
 * axis.
 */
export class LineGraph extends React.Component<LineGraphProps, {}> {
  // The SVG Element in the DOM used to render the graph.
  graphEl: SVGElement;

  // A configured NVD3 chart used to render the chart.
  chart: nvd3.LineChart;

  axis = createSelector(
    (props: {children?: React.ReactNode}) => props.children,
    (children) => {
      const axes: React.ReactElement<AxisProps>[] = findChildrenOfType(children, Axis);
      if (axes.length === 0) {
        console.warn("LineGraph requires the specification of at least one axis.");
        return null;
      }
      if (axes.length > 1) {
        console.warn("LineGraph currently only supports a single axis; ignoring additional axes.");
      }
      return axes[0];
    });

  metrics = createSelector(
    (props: {children?: React.ReactNode}) => props.children,
    (children) => {
      return findChildrenOfType(children, Metric) as React.ReactElement<MetricProps>[];
    });

  initChart() {
    const axis = this.axis(this.props);
    if (!axis) {
      // TODO: Figure out this error condition.
      return;
    }

    this.chart = nvd3.models.lineChart();
    InitLineChart(this.chart);

    if (axis.props.range) {
      this.chart.forceY(axis.props.range);
    }
  }

  mouseMove = (e: any) => {
    const { results: metrics } = this.props.data;
    if (!metrics) {
      return;
    }

    const timeScale = this.chart.xAxis.scale();

    // To get the x-coordinate within the chart we subtract the left side of the SVG
    // element and the left side margin.
    const x = e.clientX - this.graphEl.getBoundingClientRect().left - CHART_MARGINS.left;
    // Find the time value of the coordinate by asking the scale to invert the value.
    const t = Math.floor(timeScale.invert(x));

    let candidate = 0;
    let hoverTime;
    // Find first series with a successful bisect.
    while (!hoverTime && candidate < metrics.length) {
      hoverTime = bisectSeries(metrics[candidate].datapoints, t);
      candidate += 1;
    }

    if (!hoverTime) {
      return;
    }

    const positionX = e.clientX + window.scrollX;
    const positionY = e.clientY + window.scrollY;

    // Only dispatch if we have something to change to avoid action spamming.
    if (this.props.hoverState.hoverChart !== this.props.chartKey || !hoverTime.isSame(this.props.hoverState.hoverTime) || this.props.hoverState.x !== positionX || this.props.hoverState.y !== positionY) {
      this.props.hoverOn({
        hoverChart: this.props.chartKey,
        hoverTime,
        x: positionX,
        y: positionY,
      });
    }
  }

  mouseLeave = () => {
    this.props.hoverOff();
  }

  drawChart = () => {
    // If the document is not visible (e.g. if the window is minimized) we don't
    // attempt to redraw the chart. Redrawing the chart uses
    // requestAnimationFrame, which isn't called when the tab is in the
    // background, and is then apparently queued up and called en masse when the
    // tab re-enters the foreground. This check prevents the issue in #8896
    // where switching to a tab with the graphs page open that had been in the
    // background caused the UI to run out of memory and either lag or crash.
    // NOTE: This might not work on Android:
    // http://caniuse.com/#feat=pagevisibility
    if (!document.hidden) {
      const metrics = this.metrics(this.props);
      const axis = this.axis(this.props);
      if (!axis) {
        return;
      }

      const { currentlyHovering, hoverChart } = this.props.hoverState;
      let hoverTime: moment.Moment;
      let thisChart = false;
      if (currentlyHovering) {
        hoverTime = this.props.hoverState.hoverTime;
        thisChart = hoverChart !== this.props.chartKey;
      }

      ConfigureLineChart(
        this.chart, this.graphEl, metrics, axis, this.props.data, this.props.timeInfo, hoverTime, thisChart,
      );
    }
  }

  componentDidMount() {
    this.initChart();
    this.drawChart();
    // NOTE: This might not work on Android:
    // http://caniuse.com/#feat=pagevisibility
    // TODO (maxlang): Check if this element is visible based on scroll state.
    document.addEventListener("visibilitychange", this.drawChart);
  }

  componentWillUnmount() {
    document.removeEventListener("visibilitychange", this.drawChart);
  }

  componentDidUpdate() {
    this.drawChart();
  }

  render() {
    const { title, subtitle, tooltip, data } = this.props;

    return <Visualization title={title} subtitle={subtitle} tooltip={tooltip} loading={!data} >
      <div className="linegraph">
        <svg className="graph linked-guideline" ref={(svg) => this.graphEl = svg} onMouseMove={this.mouseMove} onMouseLeave={this.mouseLeave} />
      </div>
    </Visualization>;
  }
}
