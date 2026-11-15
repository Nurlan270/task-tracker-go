window.TASK_TRACKER_CONFIG = {
  API_BASE: "/api/v1/auth",
  ENDPOINTS: {
    register: { url: "/register", method: "POST" },
    login:    { url: "/login", method: "POST" },
    logout:   { url: "/logout", method: "POST" },
    refresh:  { url: "/refresh", method: "POST" },
    user:     { url: "/user", method: "GET" }
  },
  // Set the destination after successful login/register.
  AUTH_SUCCESS_URL: "/",
  // Set true if your backend requires an X-CSRF-Token header. Add token retrieval
  // according to your backend's CSRF strategy before enabling this option.
  SEND_CSRF_HEADER: false
};
