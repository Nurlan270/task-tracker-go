(function ($, window) {
  "use strict";
  const config = window.TASK_TRACKER_CONFIG;
  let refreshPromise = null;
  let sessionVersion = 0;

  function parseBody(text) {
    if (!text) return {};
    if (typeof text === "object") return text;
    try { return JSON.parse(text); } catch (_) { return {}; }
  }

  function redirectToRateLimit(xhr, request) {
    if (/^\/429\/?$/.test(window.location.pathname)) return;
    let retryAfter = xhr.getResponseHeader ? xhr.getResponseHeader("Retry-After") : null;
    let retrySeconds = 0;
    if (retryAfter) {
      const numeric = Number(retryAfter);
      retrySeconds = Number.isFinite(numeric) && numeric >= 0
        ? Math.ceil(numeric)
        : Math.max(0, Math.ceil((Date.parse(retryAfter) - Date.now()) / 1000));
    }
    const body = xhr.responseJSON || parseBody(xhr.responseText);
    try {
      sessionStorage.setItem("taskTrackerRateLimit", JSON.stringify({ retrySeconds, timestamp: Date.now(), request: request || null, returnTo: window.location.pathname + window.location.search }));
    } catch (_) {}
    window.location.replace("/429/");
  }

  function rawRequest(url, method, data) {
    const options = {
      url, method, dataType: "text", contentType: "application/json; charset=utf-8",
      xhrFields: { withCredentials: true }, headers: { Accept: "application/json" }
    };
    if (data !== undefined && method !== "GET") options.data = JSON.stringify(data);
    const d = $.Deferred();
    $.ajax(options).done(function (body, status, xhr) {
      d.resolve(parseBody(body), status, xhr);
    }).fail(function (xhr, status, error) {
      xhr.responseJSON = parseBody(xhr.responseText);
      if (xhr.status === 429) redirectToRateLimit(xhr, { url: url, method: method, data: data });
      d.reject(xhr, status, error);
    });
    return d.promise();
  }

  function authRaw(name, data) {
    const endpoint = config.ENDPOINTS[name];
    if (!endpoint) return $.Deferred().reject({ status: 0, responseJSON: { message: "" } }).promise();
    return rawRequest(config.API_BASE.replace(/\/$/, "") + endpoint.url, endpoint.method.toUpperCase(), data);
  }

  function refresh() {
    if (refreshPromise) return refreshPromise;
    const current = authRaw("refresh");
    refreshPromise = current;
    current.done(function () { sessionVersion++; });
    current.always(function () { if (refreshPromise === current) refreshPromise = null; });
    return current;
  }

  function redirectToLogin() {
    if (document.body.classList.contains("dashboard-page") && !/^\/login\/?$/.test(window.location.pathname)) {
      window.location.replace("/login/");
    }
  }

  function requestURL(url, method, data, retry, originalVersion) {
    const d = $.Deferred();
    const versionAtStart = originalVersion == null ? sessionVersion : originalVersion;
    rawRequest(url, method, data).done(function (body, status, xhr) {
      d.resolve(body, status, xhr);
    }).fail(function (xhr, status, error) {
      if (xhr.status !== 401 || retry) {
        if (xhr.status === 401 && retry) redirectToLogin();
        d.reject(xhr, status, error); return;
      }
      const refreshTask = sessionVersion !== versionAtStart
        ? $.Deferred().resolve().promise()
        : refresh();
      refreshTask.done(function () {
        requestURL(url, method, data, true, sessionVersion)
          .done(function (body, retryStatus, retryXhr) { d.resolve(body, retryStatus, retryXhr); })
          .fail(function (retryXhr, retryStatus, retryError) {
            if (retryXhr.status === 401) redirectToLogin();
            d.reject(retryXhr, retryStatus, retryError);
          });
      }).fail(function () {
        redirectToLogin();
        d.reject(xhr, status, error);
      });
    });
    return d.promise();
  }


  function authRequest(name, data) {
    if (name === "login" || name === "register" || name === "refresh") return authRaw(name, data);
    const version = sessionVersion;
    const d = $.Deferred();
    authRaw(name, data).done(function (body, status, xhr) {
      d.resolve(body, status, xhr);
    }).fail(function (xhr, status, error) {
      if (xhr.status !== 401) { d.reject(xhr, status, error); return; }
      const refreshTask = sessionVersion !== version ? $.Deferred().resolve().promise() : refresh();
      refreshTask.done(function () {
        authRaw(name, data).done(function (body, retryStatus, retryXhr) {
          d.resolve(body, retryStatus, retryXhr);
        }).fail(function (retryXhr, retryStatus, retryError) {
          retryXhr.responseJSON = parseBody(retryXhr.responseText);
          if (retryXhr.status === 401) redirectToLogin();
          d.reject(retryXhr, retryStatus, retryError);
        });
      }).fail(function () {
        redirectToLogin();
        d.reject(xhr, status, error);
      });
    });
    return d.promise();
  }

  const taskURL = config.TASKS_BASE;
  window.TaskTrackerAPI = {
    request: authRequest,
    register: data => authRequest("register", data),
    login: data => authRequest("login", data),
    logout: () => authRequest("logout"),
    refresh: refresh,
    user: () => authRequest("user"),
    tasks: {
      list: type => requestURL(taskURL + "?type=" + encodeURIComponent(type), "GET"),
      create: payload => requestURL(taskURL, "POST", payload),
      update: (id, payload) => requestURL(taskURL + encodeURIComponent(id), "PATCH", payload),
      remove: id => requestURL(taskURL + encodeURIComponent(id), "DELETE"),
      removeForce: id => requestURL(taskURL, "DELETE", { ids: [id] }),
      removeBulk: ids => requestURL(taskURL, "DELETE", { ids: ids }),
      restore: id => requestURL(taskURL + encodeURIComponent(id) + "/restore", "PATCH"),
      toggleCompletion: id => { if (id === undefined || id === null || id === "") return $.Deferred().reject({ status: 0, responseJSON: { message: "Missing task ID" } }).promise(); return requestURL(taskURL + encodeURIComponent(id) + "/toggle-completion", "PATCH"); }
    }
  };
})(jQuery, window);
