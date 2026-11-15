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

  function responseBody(xhr) {
    if (xhr.responseJSON) return xhr.responseJSON;
    if (xhr.responseText) {
      try { return JSON.parse(xhr.responseText); } catch (_) { /* Non-JSON response. */ }
    }
    return {};
  }

  function handleError(xhr, $form) {
    const body = responseBody(xhr);
    const errors = body.errors || body.validation_errors || null;
    let firstInvalid = null;

    if (errors && typeof errors === "object") {
      Object.keys(errors).forEach(function (field) {
        const message = Array.isArray(errors[field])
          ? errors[field].join(" ")
          : String(errors[field]);
        const $input = $form.find('[name="' + field + '"]');

        if ($input.length) {
          $input.attr("aria-invalid", "true").closest(".form-field").addClass("has-error");
          $form.find('[data-error-for="' + field + '"]').text(message);
          if (!firstInvalid) firstInvalid = $input;
        }
      });
    }

    // All user-facing text comes from the API response.
    showAlert(body.message, "error");
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
        .done(function (response) {
          showAlert(response && response.message, "success");
          window.setTimeout(function () {
            window.location.assign(window.TASK_TRACKER_CONFIG.AUTH_SUCCESS_URL);
          }, 450);
        })
        .fail(function (xhr) {
          if (xhr.status === 403) {
            window.location.replace("/");
            return;
          }

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
