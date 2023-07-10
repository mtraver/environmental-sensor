function constant(x) {
  return function c() {
    return x;
  };
}

function legend() {
  let rectSize = constant(12);
  let textMargin = constant(4);
  let border = false;

  function label(d) {
    return d.name;
  }

  function color(d) {
    return d.color;
  }

  function makeLegend(selection, data) {
    if (border) {
      const borderRect = selection.append("rect")
        .attr("class", "legend-border");
    }

    const leg = selection.selectAll(".legend-item")
          .data(data)
          .enter().append("g")
          .attr("class", "legend-item");

    // Add a canvas so we can use HTML5 canvas methods to get text width.
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('2d');
    context.font = selection.style('font');
    function textWidth(text) {
      return context.measureText(text).width;
    }

    // The text height is just the font size.
    const textHeight = parseInt(context.font.match(/\d+/), 10);

    let offset = 0;
    const offsets = [];
    leg.append("text")
      .attr("class", "legend-text")
      .text(label)
      .attr("x", function(d) {
        offsets.push(offset);
        const x = offset + rectSize(d) + textMargin(d);
        offset = x + textWidth(d.id) + textMargin(d) * 3;
        return x;
      })
      .attr("y", function(d) {
        return rectSize(d) / 2 + textHeight / 3;
      });

    let i = 0;
    leg.append("rect")
      .attr("class", "legend-rect")
      .attr("width", function(d) { return rectSize(d); })
      .attr("height", function(d) { return rectSize(d); })
      .attr("x", function(d) { return offsets[i++]; })
      .attr("fill", color);

    if (border) {
      const bbox = selection.node().getBBox();
      borderRect.attr("width", function(d) {
          return bbox.width + 2 * textMargin();
      }).attr("height", function(d) { return bbox.height + 2 * textMargin(); })
        .attr("x", function(d) { return bbox.x - textMargin(); })
        .attr("y", function(d) { return bbox.y - textMargin(); });
    }

    // Remove the canvas element added for text width calculation.
    canvas.remove();
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
