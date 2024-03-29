function multiExtent(data, metric, startTimestamp, endTimestamp) {
  // Get the extent of the values for each device
  const extents = data.map(d => d3.extent(
    d.values.filter(e => e.ts > startTimestamp && e.ts < endTimestamp),
    e => e[metric]));

  // Flatten the array of extents, and then get the overall extent
  return d3.extent(extents.reduce((acc, val) => acc.concat(val), []), e => e);
}

// Adds a 5% margin to the given extent (an array of length 2).
// This is useful for adding some margin to axis domains.
function padExtent(extent) {
  const adjustment = Math.abs(extent[1] - extent[0]) * 0.05;
  return [extent[0] - adjustment, extent[1] + adjustment];
}

function makePlot(selector, data, metric, name, unit, startDate, endDate) {
  // This width and height are just used here and for the viewBox (set below).
  // The SVG is given width and height of 100% in the CSS, so it scales
  // automatically based on the viewBox and preserveAspectRatio settings.
  const fullWidth = 960;
  const fullHeight = 550;

  // Margins around the focus plot (main plot) and context plot (smaller plot).
  // Vars that end with a 2 (e.g. margin2, x2) refer to the context plot.
  const margin = {top: 10, right: 20, bottom: 150, left: 65};
  const margin2 = {top: 435, right: 20, bottom: 50, left: 65};

  const svg = d3.select(selector);
  svg.attr("viewBox", "0 0 " + fullWidth + " " + fullHeight)
    .attr("preserveAspectRatio", "xMidYMid meet");

  // Calculate per-plot dimensions from the overall SVG size and the margins
  const width = fullWidth - margin.left - margin.right;
  const height = fullHeight - margin.top - margin.bottom;
  const height2 = fullHeight - margin2.top - margin2.bottom;

  const x = d3.scaleTime().range([0, width]);
  const x2 = d3.scaleTime().range([0, width]);
  const y = d3.scaleLinear().range([height, 0]);
  const y2 = d3.scaleLinear().range([height2, 0]);
  const z = d3.scaleOrdinal(d3.schemeCategory10);

  const line = d3.line()
      .curve(d3.curveBasis)
      .x(function(d) { return x(d.ts); })
      .y(function(d) { return y(d[metric]); });

  const line2 = d3.line()
      .curve(d3.curveBasis)
      .x(function(d) { return x2(d.ts); })
      .y(function(d) { return y2(d[metric]); });

  x.domain([startDate, endDate]);
  x2.domain(x.domain());

  // Filter the data based on metric, keeping only data from the devices that
  // recorded the desired metric.
  filtered = [];
  for (let i = 0; i < data.length; i++) {
    if (data[i].metrics.includes(metric)) {
      filtered.push(data[i]);
    }
  }
  data = filtered;

  // We can only set y and z domains if we have data
  if (data != null) {
    y.domain([
      d3.min(data, function(c) {
        return d3.min(c.values, function(d) {
          return d[metric];
        });
      }),
      d3.max(data, function(c) {
        return d3.max(c.values, function(d) {
          return d[metric];
        });
      })
    ]);

    // Add a 5% margin to the domain so that lines
    // aren't right at the top or bottom
    y.domain(padExtent(y.domain()));

    y2.domain(y.domain());

    z.domain(data.map(function(c) { return c.id; }));
  }

  // This tracks whether we're currently executing a zoom or brush handler in
  // order to short-circuit brush events triggered by zooms and zoom events
  // triggered by brushes. This used to be done via d3.event.sourceEvent.type,
  // but since the global d3.event was removed the sourceEvent field is no use.
  // See this issue for more info: https://github.com/d3/d3-zoom/issues/222
  let zoombrush = 0;

  const brush = d3.brushX()
      .extent([[0, 0], [width, height2]])
      .on("brush end", brushed);

  const zoom = d3.zoom()
      .scaleExtent([1, Infinity])
      .translateExtent([[0, 0], [width, height]])
      .extent([[0, 0], [width, height]])
      .on("zoom", zoomed);

  // This is used to style the axes and legend.
  const font = "17px sans-serif";

  const xAxis = d3.axisBottom(x);
  const xAxis2 = d3.axisBottom(x2);
  const yAxis = d3.axisLeft(y);

  // This ensures that the lines don't run off the plot
  const clipId = metric + "-clip";
  svg.append("defs").append("clipPath")
      .attr("id", clipId)
    .append("rect")
      .attr("width", width)
      .attr("height", height);

  // The "focus" plot is the taller one at the top
  const focus = svg.append("g")
      .attr("class", "focus")
      .attr("transform",
            "translate(" + margin.left + "," + margin.top + ")");

  // The "context" plot is the shorter one at the bottom. It doesn't have
  // a y-axis because it's so small. It's meant to give a sense of the
  // region you're looking at when you're zoomed in on the focus plot.
  const context = svg.append("g")
      .attr("class", "context")
      .attr("transform",
            "translate(" + margin2.left + "," + margin2.top + ")");

  focus.append("g")
      .attr("class", "axis axis--x")
      .style("font", font)
      .attr("transform", "translate(0," + height + ")")
      .call(xAxis);

  let yAxisLabel = name;
  if (unit != "") {
    yAxisLabel += " (" + unit + ")";
  }
  focus.append("g")
      .attr("class", "axis axis--y")
      .style("font", font)
      .call(yAxis)
    .append("text")
      .attr("text-anchor", "middle")
      .attr("transform", "rotate(-90)")
      .attr("dy", -margin.left + 12)
      .attr("dx", -height / 2)
      .attr("fill", "#000")
      .text(yAxisLabel);

  // The context plot just gets an x-axis
  context.append("g")
      .attr("class", "axis axis--x")
      .style("font", font)
      .attr("transform", "translate(0," + height2 + ")")
      .call(xAxis2);

  // If we don't have any data to display, say so and bail out
  // before attempting to draw lines
  if (data === null) {
    focus.append("g")
        .attr("transform", "translate(" + width / 2 + "," + height / 2 + ")")
      .append("text")
        .attr("text-anchor", "middle")
        .attr("fill", "#000")
        .text("No data in this time range");
    return;
  }

  // Draw lines on the focus plot
  focus.selectAll(".device")
      .data(data)
    .enter().append("g")
      .attr("class", "device")
      .attr("clip-path", "url(#" + clipId + ")")
    .append("path")
      .attr("class", "line")
      .attr("d", function(d) { return line(d.values); })
      .style("stroke", function(d) { return z(d.id); })
      .style("fill", function(d) { return "none"; });

  // Draw lines on the context plot
  context.selectAll(".device")
      .data(data)
    .enter().append("g")
      .attr("class", "device")
    .append("path")
      .attr("class", "line")
      .attr("d", function(d) { return line2(d.values); })
      .style("stroke", function(d) { return z(d.id); })
      .style("fill", function(d) { return "none"; });

  // Add a legend at the very bottom
  const legendRectSize = 15;
  const legendFunc = legend()
    .rectSize(legendRectSize)
    .label(function(d) { return d.id; })
    .color(function(d) { return z(d.id); });

  focus.append("g")
    .attr("class", "legend")
    .style("font", font)
    .attr("transform",
          "translate(" + 0 + "," + (height + margin.bottom
                                    - legendRectSize * 1.25) + ")")
    .call(legendFunc, data);

  context.append("g")
    .attr("class", "brush")
    .call(brush)
    .call(brush.move, x.range());

  svg.append("rect")
    .attr("class", "zoom")
    .attr("width", width)
    .attr("height", height)
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")")
    .call(zoom);

  // Rescales the y-axis to fit just the visible data
  function updateYAxis() {
    // Get the domain of the visible data
    let domain = multiExtent(data, metric, x.domain()[0], x.domain()[1]);

    // Add a 5% margin to the domain so that lines
    // aren't right at the top or bottom
    y.domain(padExtent(domain));

    focus.select(".axis--y").call(yAxis);
  }

  function brushed(event) {
    // Ignore brush-by-zoom
    if (zoombrush) {
      return;
    }
    zoombrush = 1;

    const s = event.selection || x2.range();
    x.domain(s.map(x2.invert, x2));
    focus.selectAll(".line")
        .attr("d", function(d) { return line(d.values); });
    focus.select(".axis--x").call(xAxis);
    svg.select(".zoom").call(
        zoom.transform,
        d3.zoomIdentity.scale(width / (s[1] - s[0])).translate(-s[0], 0));

    updateYAxis();
    zoombrush = 0;
  }

  function zoomed(event) {
    // Ignore zoom-by-brush
    if (zoombrush) {
      return;
    }
    zoombrush = 1;

    const t = event.transform;
    x.domain(t.rescaleX(x2).domain());
    focus.selectAll(".line")
        .attr("d", function(d) { return line(d.values); });
    focus.select(".axis--x").call(xAxis);
    context.select(".brush").call(brush.move, x.range().map(t.invertX, t));

    updateYAxis();
    zoombrush = 0;
  }
}
