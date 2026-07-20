/**
 * Minimal ANSI SGR → HTML converter for paste viewer.
 * Handles: reset, bold, dim, underline, fg/bg 16-color, 256-color, truecolor.
 */
(function (global) {
  "use strict";

  var ESC = "\u001b";
  var CSI_RE = /\u001b\[([0-9;]*)m/g;

  var FG = {
    30: "#000000", 31: "#ef4444", 32: "#22c55e", 33: "#eab308",
    34: "#3b82f6", 35: "#d946ef", 36: "#06b6d4", 37: "#e4e4e7",
    90: "#71717a", 91: "#f87171", 92: "#4ade80", 93: "#facc15",
    94: "#60a5fa", 95: "#e879f9", 96: "#22d3ee", 97: "#fafafa",
  };

  var BG = {
    40: "#000000", 41: "#7f1d1d", 42: "#14532d", 43: "#713f12",
    44: "#1e3a8a", 45: "#701a75", 46: "#164e63", 47: "#a1a1aa",
    100: "#3f3f46", 101: "#991b1b", 102: "#166534", 103: "#a16207",
    104: "#1d4ed8", 105: "#a21caf", 106: "#0e7490", 107: "#e4e4e7",
  };

  function ansi256(n) {
    n = Number(n);
    if (n < 0 || n > 255) return null;
    if (n < 16) {
      var basic = [
        "#000000", "#ef4444", "#22c55e", "#eab308",
        "#3b82f6", "#d946ef", "#06b6d4", "#e4e4e7",
        "#71717a", "#f87171", "#4ade80", "#facc15",
        "#60a5fa", "#e879f9", "#22d3ee", "#fafafa",
      ];
      return basic[n];
    }
    if (n < 232) {
      n -= 16;
      var r = Math.floor(n / 36);
      var g = Math.floor((n % 36) / 6);
      var b = n % 6;
      var steps = [0, 95, 135, 175, 215, 255];
      return "rgb(" + steps[r] + "," + steps[g] + "," + steps[b] + ")";
    }
    var gray = 8 + (n - 232) * 10;
    return "rgb(" + gray + "," + gray + "," + gray + ")";
  }

  function escapeHtml(text) {
    return text
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function styleToCss(style) {
    var parts = [];
    if (style.bold) parts.push("font-weight:700");
    if (style.dim) parts.push("opacity:0.65");
    if (style.underline) parts.push("text-decoration:underline");
    if (style.fg) parts.push("color:" + style.fg);
    if (style.bg) parts.push("background-color:" + style.bg);
    return parts.join(";");
  }

  function applySgr(style, codes) {
    if (!codes.length || (codes.length === 1 && codes[0] === "")) {
      codes = ["0"];
    }

    for (var i = 0; i < codes.length; i++) {
      var c = Number(codes[i]);
      if (c === 0 || isNaN(c)) {
        style.bold = false;
        style.dim = false;
        style.underline = false;
        style.fg = null;
        style.bg = null;
      } else if (c === 1) {
        style.bold = true;
      } else if (c === 2) {
        style.dim = true;
      } else if (c === 4) {
        style.underline = true;
      } else if (c === 22) {
        style.bold = false;
        style.dim = false;
      } else if (c === 24) {
        style.underline = false;
      } else if (c === 39) {
        style.fg = null;
      } else if (c === 49) {
        style.bg = null;
      } else if (FG[c]) {
        style.fg = FG[c];
      } else if (BG[c]) {
        style.bg = BG[c];
      } else if (c === 38 || c === 48) {
        var isFg = c === 38;
        var mode = Number(codes[++i]);
        if (mode === 5) {
          var color = ansi256(codes[++i]);
          if (isFg) style.fg = color;
          else style.bg = color;
        } else if (mode === 2) {
          var r = Number(codes[++i]);
          var g = Number(codes[++i]);
          var b = Number(codes[++i]);
          var rgb = "rgb(" + r + "," + g + "," + b + ")";
          if (isFg) style.fg = rgb;
          else style.bg = rgb;
        }
      }
    }
  }

  function hasAnsi(text) {
    return text.indexOf(ESC + "[") !== -1;
  }

  function ansiToHtml(text) {
    if (!text) return "";

    var style = {
      bold: false,
      dim: false,
      underline: false,
      fg: null,
      bg: null,
    };

    var html = "";
    var last = 0;
    var match;

    CSI_RE.lastIndex = 0;
    while ((match = CSI_RE.exec(text)) !== null) {
      var chunk = text.slice(last, match.index);
      if (chunk) {
        var css = styleToCss(style);
        var escaped = escapeHtml(chunk);
        html += css
          ? '<span style="' + css + '">' + escaped + "</span>"
          : escaped;
      }
      applySgr(style, match[1].split(";"));
      last = match.index + match[0].length;
    }

    var tail = text.slice(last);
    if (tail) {
      var tailCss = styleToCss(style);
      var tailEsc = escapeHtml(tail);
      html += tailCss
        ? '<span style="' + tailCss + '">' + tailEsc + "</span>"
        : tailEsc;
    }

    return html;
  }

  global.AnsiUp = {
    hasAnsi: hasAnsi,
    ansiToHtml: ansiToHtml,
  };
})(typeof window !== "undefined" ? window : this);
