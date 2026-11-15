
(function ($, window, document) {
  "use strict";

  function parseResponse(raw) {
    if (!raw) return {};

    if (typeof raw === "object") return raw;

    try {
      return JSON.parse(raw);
    } catch (_) {
      return {};
    }
  }

  function getUser(response) {
    return response.user ||
        response.data?.user ||
        response.data ||
        response;
  }

  function isAuthPage() {
    return /^\/(login|register)\/?$/.test(window.location.pathname);
  }

  function renderAuthenticatedHeader(user) {
    const $actions = $(".header-actions");
    if (!$actions.length) return;

    const $wrapper = $("<div>", {
      class: "user-dropdown"
    });

    const $trigger = $("<button>", {
      type: "button",
      class: "user-trigger",
      "aria-expanded": "false"
    });

    $("<span>", {
      class: "user-avatar",
      text: (user.name || "?").charAt(0).toUpperCase()
    }).appendTo($trigger);

    $("<span>", {
      class: "user-name",
      text: user.name
    }).appendTo($trigger);

    $("<span>", {
      class: "dropdown-chevron",
      text: "▾"
    }).appendTo($trigger);

    const $menu = $("<div>", {
      class: "user-menu",
      hidden: true
    });

    $("<div>", {
      class: "user-menu-name",
      text: user.name
    }).appendTo($menu);

    $("<div>", {
      class: "user-menu-email",
      text: user.email
    }).appendTo($menu);

    $("<div>", {
      class: "user-menu-divider"
    }).appendTo($menu);

    const $logout = $("<button>", {
      type: "button",
      class: "logout-button",
      text: "Log out"
    });

    const $logoutMessage = $("<div>", {
      class: "logout-message",
      role: "alert"
    });

    $menu.append($logout, $logoutMessage);
    $wrapper.append($trigger, $menu);
    $actions.empty().append($wrapper);

    $trigger.on("click", function () {
      const isOpen = !$menu.prop("hidden");

      $menu.prop("hidden", isOpen);
      $trigger.attr("aria-expanded", String(!isOpen));
    });

    $(document).on("click", function (event) {
      if (!$(event.target).closest(".user-dropdown").length) {
        $menu.prop("hidden", true);
        $trigger.attr("aria-expanded", "false");
      }
    });

    $(document).on("keydown", function (event) {
      if (event.key === "Escape") {
        $menu.prop("hidden", true);
        $trigger.attr("aria-expanded", "false");
      }
    });

    $logout.on("click", function () {
      $logout.prop("disabled", true);
      $logoutMessage.text("");

      window.TaskTrackerAPI.logout()
          .done(function () {
            window.location.assign("/");
          })
          .fail(function (xhr) {
            const response = parseResponse(xhr.responseText);
            $logoutMessage.text(response.message || "");
            $logout.prop("disabled", false);
          });
    });
  }

  function init() {
    $(".year, #year").text(new Date().getFullYear());

    const header = document.getElementById("site-header");

    if (header) {
      const updateHeader = () => {
        header.classList.toggle("is-scrolled", window.scrollY > 8);
      };

      updateHeader();
      window.addEventListener("scroll", updateHeader, { passive: true });
    }

    // Check whether the current visitor is authenticated.
    window.TaskTrackerAPI.user()
        .done(function (rawResponse) {
          const response = parseResponse(rawResponse);
          const user = getUser(response);

          if (isAuthPage()) {
            window.location.replace("/");
            return;
          }

          if (!user || !user.name) return;

          renderAuthenticatedHeader(user);

          // Authenticated users see an empty main page for now.
          $(".home-main .welcome-card").remove();
        })
        .fail(function (xhr) {
          // The backend may use 403 for authenticated users accessing guest pages.
          if (isAuthPage() && xhr.status === 403) {
            window.location.replace("/");
          }
        });
  }

  $(init);

})(jQuery, window, document);
