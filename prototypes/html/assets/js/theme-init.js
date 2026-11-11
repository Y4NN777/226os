/* Runs synchronously in <head> to avoid a theme flash. ?theme=dark|light wins, then localStorage; otherwise OS preference (CSS). */
(function () {
  try {
    var t = new URLSearchParams(location.search).get('theme');
    if (t !== 'dark' && t !== 'light') t = localStorage.getItem('226os.theme');
    if (t === 'dark' || t === 'light') document.documentElement.setAttribute('data-theme', t);
  } catch (e) {}
})();
