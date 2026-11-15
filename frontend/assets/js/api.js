
(function ($, window) {
  "use strict";

  const config = window.TASK_TRACKER_CONFIG;

  let refreshPromise = null;
  let sessionVersion = 0;
  let refreshRejected = false;

  function parseResponse(body) {
    if (!body) return {};

    if (typeof body === "object") return body;

    try {
      return JSON.parse(body);
    } catch (_) {
      return {};
    }
  }

  // Sends one HTTP request without automatic refresh handling.
  function send(endpointName, data) {
    const endpoint = config.ENDPOINTS[endpointName];

    if (!endpoint) {
      const xhr = {
        status: 0,
        responseJSON: {}
      };

      return $.Deferred()
          .reject(xhr, "error")
          .promise();
    }

    const method = endpoint.method.toUpperCase();

    const options = {
      url: config.API_BASE.replace(/\/$/, "") + endpoint.url,
      method: method,
      dataType: "text",
      headers: {
        Accept: "application/json"
      },
      xhrFields: {
        withCredentials: true
      }
    };

    if (data !== undefined && method !== "GET") {
      options.contentType = "application/json; charset=utf-8";
      options.data = JSON.stringify(data);
    }

    const deferred = $.Deferred();

    $.ajax(options)
        .done(function (body, textStatus, xhr) {
          deferred.resolve(
              parseResponse(body),
              textStatus,
              xhr
          );
        })
        .fail(function (xhr, textStatus, error) {
          xhr.responseJSON = parseResponse(xhr.responseText);

          deferred.reject(xhr, textStatus, error);
        });

    return deferred.promise();
  }

  // Shared refresh request: concurrent 401 responses
  // reuse the same promise instead of refreshing multiple times.
  function refreshSession() {
    if (refreshRejected) {
      return $.Deferred()
          .reject({
            status: 401,
            responseJSON: {}
          }, "error")
          .promise();
    }

    if (refreshPromise) {
      return refreshPromise;
    }

    const currentPromise = send("refresh");

    refreshPromise = currentPromise;

    currentPromise.done(function () {
      sessionVersion++;
      refreshRejected = false;
    });

    currentPromise.fail(function (xhr) {
      // Don't repeatedly attempt an invalid/expired refresh token.
      if (xhr.status === 401) {
        refreshRejected = true;
      }
    });

    currentPromise.always(function () {
      if (refreshPromise === currentPromise) {
        refreshPromise = null;
      }
    });

    return currentPromise;
  }

  // Makes an API request and automatically refreshes on 401.
  function request(endpointName, data, alreadyRetried) {
    const deferred = $.Deferred();
    const requestVersion = sessionVersion;

    const excludedFromRefresh = [
      "login",
      "register",
      "refresh"
    ];

    send(endpointName, data)
        .done(function (response, textStatus, xhr) {
          // Successful authentication changes the session.
          if (
              endpointName === "login" ||
              endpointName === "register"
          ) {
            sessionVersion++;
            refreshRejected = false;
          }

          if (endpointName === "logout") {
            sessionVersion++;
            refreshRejected = true;
          }

          deferred.resolve(response, textStatus, xhr);
        })
        .fail(function (xhr, textStatus, error) {
          const shouldRefresh =
              xhr.status === 401 &&
              !alreadyRetried &&
              !excludedFromRefresh.includes(endpointName);

          if (!shouldRefresh) {
            deferred.reject(xhr, textStatus, error);
            return;
          }

          // Another request may already have refreshed
          // the session while this request was in flight.
          const refresh = sessionVersion !== requestVersion
              ? $.Deferred().resolve().promise()
              : refreshSession();

          refresh
              .done(function () {
                // Retry the original request only once.
                request(endpointName, data, true)
                    .done(function (response, status, retryXhr) {
                      deferred.resolve(response, status, retryXhr);
                    })
                    .fail(function (retryXhr, retryStatus, retryError) {
                      deferred.reject(
                          retryXhr,
                          retryStatus,
                          retryError
                      );
                    });
              })
              .fail(function (refreshXhr, refreshStatus, refreshError) {
                deferred.reject(
                    refreshXhr || xhr,
                    refreshStatus || textStatus,
                    refreshError || error
                );
              });
        });

    return deferred.promise();
  }

  window.TaskTrackerAPI = {
    request: request,

    register: function (data) {
      return request("register", data);
    },

    login: function (data) {
      return request("login", data);
    },

    logout: function () {
      return request("logout");
    },

    refresh: refreshSession,

    user: function () {
      return request("user");
    }
  };

})(jQuery, window);
