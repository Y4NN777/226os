/* Runs synchronously in <head> to avoid a flash. Theme: ?theme=dark|light wins, then localStorage ('system' or nothing = OS preference via CSS).
   Prototype annotations: ?proto=1|0 wins, then localStorage 226os.proto.annotations; adds html.proto-annotations (shows every [data-proto]). */
(function () {
  try {
    var P = new URLSearchParams(location.search), t = P.get('theme'), root = document.documentElement;
    if (t !== 'dark' && t !== 'light') t = localStorage.getItem('226os.theme');
    if (t === 'dark' || t === 'light') root.setAttribute('data-theme', t);
    var a = P.get('proto'); if (a !== '1' && a !== '0') a = localStorage.getItem('226os.proto.annotations');
    if (a === '1') root.classList.add('proto-annotations');
  } catch (e) {}
})();
