function multiExtent(data, startTimestamp, endTimestamp) {
  // Get the extent of the values for each device
  var extents = data.map(d => d3.extent(
    d.values.filter(e => e.ts > startTimestamp && e.ts < endTimestamp),
    e => e.t));

  // Flatten the array of extents, and then get the overall extent
  return d3.extent(extents.reduce((acc, val) => acc.concat(val), []), e => e);
}

// Adds a 5% margin to the given extent (an array of length 2).
// This is useful for adding some margin to axis domains.
function padExtent(extent) {
  var adjustment = Math.abs(extent[1] - extent[0]) * 0.05;
  return [extent[0] - adjustment, extent[1] + adjustment];
}

function makePlot(selector, data, startDate, endDate) {
  // This width and height are just used here and for the viewBox (set below).
  // The SVG is given width and height of 100% in the CSS, so it scales
  // automatically based on the viewBox and preserveAspectRatio settings.
  var fullWidth = 960;
  var fullHeight = 550;

  // Margins around the focus plot (main plot) and context plot (smaller plot).
  // Vars that end with a 2 (e.g. margin2, x2) refer to the context plot.
  var margin = {top: 5, right: 20, bottom: 150, left: 50};
  var margin2 = {top: 435, right: 20, bottom: 50, left: 50};

  var svg = d3.select(selector);
  svg.attr("viewBox", "0 0 " + fullWidth + " " + fullHeight)
    .attr("preserveAspectRatio", "xMidYMid meet");

  // Calculate per-plot dimensions from the overall SVG size and the margins
  var width = fullWidth - margin.left - margin.right;
  var height = fullHeight - margin.top - margin.bottom;
  var height2 = fullHeight - margin2.top - margin2.bottom;

  var x = d3.scaleTime().range([0, width]);
  var x2 = d3.scaleTime().range([0, width]);
  var y = d3.scaleLinear().range([height, 0]);
  var y2 = d3.scaleLinear().range([height2, 0]);
  var z = d3.scaleOrdinal(d3.schemeCategory10);

  var line = d3.line()
      .curve(d3.curveBasis)
      .x(function(d) { return x(d.ts); })
      .y(function(d) { return y(d.t); });

  var line2 = d3.line()
      .curve(d3.curveBasis)
      .x(function(d) { return x2(d.ts); })
      .y(function(d) { return y2(d.t); });

  x.domain([startDate, endDate]);
  x2.domain(x.domain());

  // We can only set y and z domains if we have data
  if (data != null) {
    y.domain([
      d3.min(data, function(c) {
        return d3.min(c.values, function(d) {
          return d.t;
        });
      }),
      d3.max(data, function(c) {
        return d3.max(c.values, function(d) {
          return d.t;
        });
      })
    ]);

    // Add a 5% margin to the domain so that lines
    // aren't right at the top or bottom
    y.domain(padExtent(y.domain()));

    y2.domain(y.domain());

    z.domain(data.map(function(c) { return c.id; }));
  }

  var brush = d3.brushX()
      .extent([[0, 0], [width, height2]])
      .on("brush end", brushed);

  var zoom = d3.zoom()
      .scaleExtent([1, Infinity])
      .translateExtent([[0, 0], [width, height]])
      .extent([[0, 0], [width, height]])
      .on("zoom", zoomed);

  var xAxis = d3.axisBottom(x);
  var xAxis2 = d3.axisBottom(x2);
  var yAxis = d3.axisLeft(y);

  // This ensures that the lines don't run off the plot
  svg.append("defs").append("clipPath")
      .attr("id", "clip")
    .append("rect")
      .attr("width", width)
      .attr("height", height);

  // The "focus" plot is the taller one at the top
  var focus = svg.append("g")
      .attr("class", "focus")
      .attr("transform",
            "translate(" + margin.left + "," + margin.top + ")");

  // The "context" plot is the shorter one at the bottom. It doesn't have
  // a y-axis because it's so small. It's meant to give a sense of the
  // region you're looking at when you're zoomed in on the focus plot.
  var context = svg.append("g")
      .attr("class", "context")
      .attr("transform",
            "translate(" + margin2.left + "," + margin2.top + ")");

  focus.append("g")
      .attr("class", "axis axis--x")
      .attr("transform", "translate(0," + height + ")")
      .call(xAxis);

  focus.append("g")
      .attr("class", "axis axis--y")
      .call(yAxis)
    .append("text")
      .attr("text-anchor", "middle")
      .attr("transform", "rotate(-90)")
      .attr("dy", -margin.left + 15)
      .attr("dx", -height / 2)
      .attr("fill", "#000")
      .text("Temperature (°C)");

  // The context plot just gets an x-axis
  context.append("g")
      .attr("class", "axis axis--x")
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
  var legendRectSize = 15;
  var legendFunc = legend()
    .rectSize(legendRectSize)
    .label(function(d) { return d.id; })
    .color(function(d) { return z(d.id); });

  focus.append("g")
    .attr("class", "legend")
    .style("font", "11px sans-serif")
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
    var domain = multiExtent(data, x.domain()[0], x.domain()[1]);

    // Add a 5% margin to the domain so that lines
    // aren't right at the top or bottom
    y.domain(padExtent(domain));

    focus.select(".axis--y").call(yAxis);
  }

  function brushed() {
    if (d3.event.sourceEvent && d3.event.sourceEvent.type === "zoom") {
      // Ignore brush-by-zoom
      return;
    }

    var s = d3.event.selection || x2.range();
    x.domain(s.map(x2.invert, x2));
    focus.selectAll(".line")
        .attr("d", function(d) { return line(d.values); });
    focus.select(".axis--x").call(xAxis);
    svg.select(".zoom").call(
        zoom.transform,
        d3.zoomIdentity.scale(width / (s[1] - s[0])).translate(-s[0], 0));

    updateYAxis();
  }

  function zoomed() {
    if (d3.event.sourceEvent && d3.event.sourceEvent.type === "brush") {
      // Ignore zoom-by-brush
      return;
    }

    var t = d3.event.transform;
    x.domain(t.rescaleX(x2).domain());
    focus.selectAll(".line")
        .attr("d", function(d) { return line(d.values); });
    focus.select(".axis--x").call(xAxis);
    context.select(".brush").call(brush.move, x.range().map(t.invertX, t));

    updateYAxis();
  }
}
