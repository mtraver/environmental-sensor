function constant(x) {
  return function c() {
    return x;
  };
}

function legend() {
  var rectSize = constant(12);
  var textMargin = constant(4);
  var border = false;

  function label(d) {
    return d.name;
  }

  function color(d) {
    return d.color;
  }

  function makeLegend(selection, data) {
    if (border) {
      var borderRect = selection.append("rect")
        .attr("class", "legend-border");
    }

    var leg = selection.selectAll(".legend-item")
          .data(data)
          .enter().append("g")
          .attr("class", "legend-item");

    var offset = 0;
    var offsets = [];
    leg.append("text")
      .attr("class", "legend-text")
      .text(label)
      .attr("x", function(d) {
        offsets.push(offset);
        var x = offset + rectSize(d) + textMargin(d);
        offset = x + this.getBBox().width + textMargin(d) * 3;
        return x;
      })
      .attr("y", function(d) {
        return rectSize(d) / 2 + this.getBBox().height / 3;
      });

    var i = 0;
    leg.append("rect")
      .attr("class", "legend-rect")
      .attr("width", function(d) { return rectSize(d); })
      .attr("height", function(d) { return rectSize(d); })
      .attr("x", function(d) { return offsets[i++]; })
      .attr("fill", color);

    if (border) {
      var bbox = selection.node().getBBox();
      borderRect.attr("width", function(d) {
          return bbox.width + 2 * textMargin();
      }).attr("height", function(d) { return bbox.height + 2 * textMargin(); })
        .attr("x", function(d) { return bbox.x - textMargin(); })
        .attr("y", function(d) { return bbox.y - textMargin(); });
    }
  }

  makeLegend.rectSize = function(s) {
    return arguments.length ? (
        rectSize = typeof s === "function" ? s : constant(+s),
        makeLegend) : rectSize;
  }

  makeLegend.textMargin = function(m) {
    return arguments.length ? (
        textMargin = typeof m === "function" ? m : constant(+m),
        makeLegend) : textMargin;
  }

  makeLegend.border = function(b) {
    return arguments.length ? (border = b, makeLegend) : border;
  }

  makeLegend.label = function(l) {
    return arguments.length ? (
        label = typeof l === "function" ? l : constant(+l), makeLegend) : label;
  }

  makeLegend.color = function(c) {
    return arguments.length ? (
        color = typeof c === "function" ? c : constant(+c), makeLegend) : color;
  }

  return makeLegend;
}
