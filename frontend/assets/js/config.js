window.TASK_TRACKER_CONFIG = {
  API_BASE: "/api/v1/auth",
  TASKS_BASE: "/api/v1/tasks/",
  ENDPOINTS: {
    register: { url: "/register", method: "POST" },
    login:    { url: "/login", method: "POST" },
    logout:   { url: "/logout", method: "POST" },
    refresh:  { url: "/refresh", method: "POST" },
    user:     { url: "/user", method: "GET" }
  },
  AUTH_SUCCESS_URL: "/",
  SEND_CSRF_HEADER: false
};
