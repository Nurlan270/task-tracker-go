(function ($, window) {
  "use strict";

  function clearErrors($form) {
    $form.find(".field-error").text("");
    $form.find(".form-field").removeClass("has-error");
    $form.find("input").attr("aria-invalid", "false");
    $("#form-alert").prop("hidden", true).removeClass("is-error is-success").text("");
  }

  function showAlert(message, type) {
    if (!message) return;
    $("#form-alert")
      .removeClass("is-error is-success")
      .addClass(type === "success" ? "is-success" : "is-error")
      .text(message)
      .prop("hidden", false);
  }

  function parseResponse(raw) {
    if (!raw) return {};
    try {
      return JSON.parse(raw);
    } catch (_) {
      return {};
    }
  }

  function responseBody(xhr) {
    if (xhr.responseJSON) return xhr.responseJSON;
    return parseResponse(xhr.responseText);
  }

  function flattenErrors(value) {
    if (value == null) return [];
    if (Array.isArray(value)) return value.flatMap(flattenErrors);
    if (typeof value === "object") return Object.values(value).flatMap(flattenErrors);
    return [String(value)];
  }

  function handleError(xhr, $form) {
    const body = responseBody(xhr);
    const errors = body.errors || body.validation_errors || null;
    let firstInvalid = null;
    const errorEntries = errors && typeof errors === "object" ? Object.entries(errors) : [];

    if (errorEntries.length) {
      errorEntries.forEach(function ([field, value]) {
        const messages = flattenErrors(value).join(" ");
        const $input = $form.find('[name="' + field + '"]');
        if ($input.length) {
          $input.attr("aria-invalid", "true").closest(".form-field").addClass("has-error");
          $form.find('[data-error-for="' + field + '"]').text(messages);
          if (!firstInvalid) firstInvalid = $input;
        } else {
          showAlert(messages, "error");
        }
      });
      // Validation errors take precedence over the generic API message.
      if (!firstInvalid && !errorEntries.some(([field]) => $form.find('[name="' + field + '"]').length === 0)) {
        $("#form-alert").prop("hidden", true).text("");
      }
    } else {
      showAlert(body.message, "error");
    }

    if (firstInvalid) firstInvalid.trigger("focus");
  }

  function setLoading($form, loading) {
    const $button = $form.find('button[type="submit"]');
    $button.prop("disabled", loading).toggleClass("is-loading", loading);
  }

  function bindForm(selector, action) {
    $(selector).on("submit", function (event) {
      event.preventDefault();
      const $form = $(this);
      clearErrors($form);
      if (!$form[0].reportValidity()) return;

      const payload = {};
      $form.serializeArray().forEach(item => { payload[item.name] = item.value; });

      setLoading($form, true);
      window.TaskTrackerAPI[action](payload)
        .done(function (rawResponse) {
          const response = parseResponse(rawResponse);
          showAlert(response.message, "success");
          window.setTimeout(function () {
            window.location.assign(window.TASK_TRACKER_CONFIG.AUTH_SUCCESS_URL);
          }, 450);
        })
        .fail(function (xhr) {
          handleError(xhr, $form);
        })
        .always(function () {
          setLoading($form, false);
        });
    });
  }

  $(function () {
    bindForm("#login-form", "login");
    bindForm("#register-form", "register");
  });
})(jQuery, window);
