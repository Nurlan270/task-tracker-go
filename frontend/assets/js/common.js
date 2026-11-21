(function ($, window, document) {
  "use strict";
  function parseResponse(raw) {
    if (!raw) return {};
    if (typeof raw === "object") return raw;
    try { return JSON.parse(raw); } catch (_) { return {}; }
  }
  function unwrapUser(response) {
    return response.user || response.data?.user || response.data || response;
  }
  function isAuthRoute() { return /^\/(login|register)\/?$/.test(location.pathname); }

  function renderUserMenu(user) {
    const $actions = $(".header-actions");
    if (!$actions.length) return;
    const $wrap = $("<div class='user-dropdown'></div>");
    const $trigger = $("<button type='button' class='user-trigger' aria-expanded='false'></button>");
    $trigger.append($('<span class="user-avatar"></span>').text((user.name || "?").charAt(0).toUpperCase()));
    $trigger.append($('<span class="user-name"></span>').text(user.name || ""));
    $trigger.append("<span class='dropdown-chevron' aria-hidden='true'>▾</span>");
    const $menu = $("<div class='user-menu' hidden></div>");
    $menu.append($('<div class="user-menu-name"></div>').text(user.name || ""));
    $menu.append($('<div class="user-menu-email"></div>').text(user.email || ""));
    $menu.append("<div class='user-menu-divider'></div>");
    const $logout = $("<button type='button' class='logout-button'>Logout</button>");
    const $message = $("<p class='logout-message' role='alert'></p>");
    $menu.append($logout, $message);
    $wrap.append($trigger, $menu);
    $actions.empty().append($wrap);
    $trigger.on("click", function (event) {
      event.stopPropagation();
      const open = !$menu.prop("hidden");
      $menu.prop("hidden", open);
      $trigger.attr("aria-expanded", String(!open));
    });
    $(document).on("click.taskTrackerMenu", function (event) {
      if (!$(event.target).closest(".user-dropdown").length) {
        $menu.prop("hidden", true);
        $trigger.attr("aria-expanded", "false");
      }
    }).on("keydown.taskTrackerMenu", function (event) {
      if (event.key === "Escape") { $menu.prop("hidden", true); $trigger.attr("aria-expanded", "false"); }
    });
    $logout.on("click", function () {
      $logout.prop("disabled", true).text("...");
      window.TaskTrackerAPI.logout().done(function () {
        window.location.assign("/");
      }).fail(function (xhr) {
        const response = xhr.responseJSON || parseResponse(xhr.responseText);
        $message.text(response.message || "");
        $logout.prop("disabled", false).text("Logout");
      });
    });
  }

  $(function () {
    $("#year, .year").text(new Date().getFullYear());
    const header = document.getElementById("site-header");
    if (header) {
      const updateHeader = () => header.classList.toggle("is-scrolled", window.scrollY > 8);
      updateHeader();
      window.addEventListener("scroll", updateHeader, { passive: true });
    }
    window.TaskTrackerSession = window.TaskTrackerAPI.user();
    window.TaskTrackerSession.done(function (raw) {
      const user = unwrapUser(parseResponse(raw));
      if (isAuthRoute()) { window.location.replace("/"); return; }
      if (!user || !user.name) return;
      renderUserMenu(user);
    }).fail(function (xhr) {
      if (isAuthRoute() && xhr.status === 403) window.location.replace("/");
      // Dashboard auth is gated by tasks.js; never render guest buttons over a protected page.
    });
  });
})(jQuery, window, document);
