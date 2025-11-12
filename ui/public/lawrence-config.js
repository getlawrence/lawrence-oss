(function () {
  const placeholder = "__LAWRENCE_BACKEND_URL__";
  if (typeof window === "undefined") {
    return;
  }

  window.__LAWRENCE_CONFIG__ = window.__LAWRENCE_CONFIG__ ?? {};

  if (!window.__LAWRENCE_CONFIG__.backendUrl) {
    window.__LAWRENCE_CONFIG__.backendUrl = placeholder;
  }
})();

