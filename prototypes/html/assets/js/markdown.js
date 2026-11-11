/* Shared tiny Markdown renderer (no deps, no network). Escapes ALL input first, then supports headings, lists, quotes, fenced code
   (with a small highlighter + copy button), inline code/bold/italic and http(s)/relative links. Used by Solve, Resource and Project.
   Markdown.render(md, {id, headStart, copyText, copyLabel(lang), codeLabel(lang)})  headStart = level of "#" (default 3; use 2 so "##" -> h3).
   Markdown.unbalancedFence(md) -> true when a ``` fence is opened but never closed. Load with <script defer> before page scripts. */
(function () {
  'use strict';
  var esc = function (s) { return String(s).replace(/[&<>"']/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]; }); };
  var Markdown = {
    hl: function (code, lang) {
      var kw = /^(func|package|import|return|if|else|for|range|var|const|let|type|struct|defer|nil|true|false|async|await|function|new|select|from|where|create|table|insert|into|values|on|conflict|do|nothing|primary|key|not|null|unique|index|using|alter|add|constraint|references)$/i;
      var com = { go: '//', js: '//', ts: '//', php: '//', sql: '--', python: '#', sh: '#', bash: '#', yaml: '#' }[lang] || '//';
      var re = new RegExp('(' + com.replace(/[/#-]/g, '\\$&') + '[^\\n]*)|("(?:\\\\.|[^"\\\\\\n])*"|\'(?:\\\\.|[^\'\\\\\\n])*\')|([A-Za-z_][A-Za-z0-9_]*)(\\()?', 'g');
      var out = '', last = 0, m;
      while ((m = re.exec(code))) {
        out += esc(code.slice(last, m.index)); last = re.lastIndex;
        if (m[1]) out += '<span class="tok-com">' + esc(m[1]) + '</span>';
        else if (m[2]) out += '<span class="tok-str">' + esc(m[2]) + '</span>';
        else if (kw.test(m[3]) && !m[4]) out += '<span class="tok-kw">' + esc(m[3]) + '</span>';
        else if (m[4]) out += '<span class="tok-fn">' + esc(m[3]) + '</span>' + esc(m[4]);
        else out += esc(m[3]);
      }
      return out + esc(code.slice(last));
    },
    inline: function (s) {
      s = esc(s);
      s = s.replace(/`([^`]+)`/g, '<code>$1</code>');
      s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>').replace(/(^|[^*])\*([^*\s][^*]*)\*/g, '$1<em>$2</em>');
      s = s.replace(/\[([^\]]+)\]\(((?:https?:\/\/|\/|\.\.?\/)[^)\s]+)\)/g, function (m, t, u) {
        return /^https?:/.test(u) ? '<a class="link-external" href="' + u + '" rel="noopener noreferrer">' + t + '</a>' : '<a href="' + u + '">' + t + '</a>';
      });
      return s;
    },
    /* opts: {id: unique prefix for code ids/testids, copyLabel(lang), codeLabel(lang), headStart: first heading level} */
    render: function (md, o) {
      o = o || {}; var lines = String(md || '').replace(/\r/g, '').split('\n'), html = '', i = 0, n = 0, para = [], list = null;
      var inl = Markdown.inline;
      function flushP() { if (para.length) { html += '<p>' + inl(para.join(' ')) + '</p>'; para = []; } }
      function flushL() { if (list) { html += '<' + list.t + '>' + list.items.map(function (x) { return '<li>' + inl(x) + '</li>'; }).join('') + '</' + list.t + '>'; list = null; } }
      while (i < lines.length) {
        var ln = lines[i], f = ln.match(/^(```|~~~)\s*([A-Za-z0-9+#-]*)\s*$/);
        if (f) {
          flushP(); flushL(); var buf = []; i++;
          while (i < lines.length && !/^(```|~~~)\s*$/.test(lines[i])) { buf.push(lines[i]); i++; }
          i++; n++; var lang = (f[2] || 'text').toLowerCase(), cid = (o.id || 'md') + '-code-' + n, name = { go: 'Go', sql: 'SQL', php: 'PHP', ts: 'TypeScript', js: 'JavaScript', python: 'Python', sh: 'Shell', bash: 'Shell' }[lang] || lang;
          html += '<div class="code-block" data-testid="' + cid + '-block"><div class="code-head"><span data-testid="' + cid + '-lang">' + esc(name) + '</span>' +
            '<button type="button" class="btn btn-sm" data-copy="#' + cid + '" aria-label="' + esc(o.copyLabel ? o.copyLabel(name) : 'Copy') + '" data-testid="' + cid + '-copy"><span class="icon i-copy" aria-hidden="true"></span><span>' + esc(o.copyText || 'Copy') + '</span></button></div>' +
            '<pre tabindex="0" aria-label="' + esc(o.codeLabel ? o.codeLabel(name) : name) + '" data-testid="' + cid + '-pre"><code id="' + cid + '">' + Markdown.hl(buf.join('\n'), lang) + '</code></pre></div>';
          continue;
        }
        var h = ln.match(/^(#{1,4})\s+(.*)$/), li = ln.match(/^\s*[-*]\s+(.*)$/), ol = ln.match(/^\s*\d+[.)]\s+(.*)$/), bq = ln.match(/^>\s?(.*)$/);
        if (!ln.trim()) { flushP(); flushL(); }
        else if (h) { flushP(); flushL(); var lv = Math.min(4, (o.headStart || 3) + h[1].length - 1); html += '<h' + lv + '>' + inl(h[2]) + '</h' + lv + '>'; }
        else if (li) { flushP(); if (!list || list.t !== 'ul') { flushL(); list = { t: 'ul', items: [] }; } list.items.push(li[1]); }
        else if (ol) { flushP(); if (!list || list.t !== 'ol') { flushL(); list = { t: 'ol', items: [] }; } list.items.push(ol[1]); }
        else if (bq) { flushP(); flushL(); html += '<blockquote><p>' + inl(bq[1]) + '</p></blockquote>'; }
        else { flushL(); para.push(ln.trim()); }
        i++;
      }
      flushP(); flushL(); return html;
    },
    /* true when a ``` fence is opened but never closed (client-side "snippet valide" check) */
    unbalancedFence: function (md) { var c = (String(md).match(/^(```|~~~)/gm) || []).length; return c % 2 === 1; }
  };
  window.Markdown = Markdown;
})();
