(function ($, window, document) {
  "use strict";

  const types = ["uncompleted", "completed", "deleted"];
  const titles = {
    uncompleted: "Uncompleted Tasks",
    completed: "Completed Tasks",
    deleted: "Trash"
  };
  const subtitles = {
    uncompleted: "Tasks you still need to complete",
    completed: "Tasks you have completed",
    deleted: "Deleted tasks"
  };
  const cache = { uncompleted: null, completed: null, deleted: null };
  let currentType = "uncompleted";
  let selectedTaskId = null;
  let isLoading = false;
  let animateNextList = false;
  const selectedForBulk = new Set();

  function parseResponse(raw) {
    if (!raw) return {};
    if (typeof raw === "object") return raw;
    try { return JSON.parse(raw); } catch (_) { return {}; }
  }
  function unwrapList(response) {
    if (Array.isArray(response)) return response;
    if (Array.isArray(response.tasks)) return response.tasks;
    if (Array.isArray(response.data)) return response.data;
    if (Array.isArray(response.data?.tasks)) return response.data.tasks;
    if (Array.isArray(response.items)) return response.items;
    if (Array.isArray(response.results)) return response.results;
    if (Array.isArray(response.data?.items)) return response.data.items;
    return [];
  }
  // The list endpoint contract is authoritative: id, title, description, created_at, completed_at.
  function normalize(task) {
    const normalized = {
      id: task.id,
      title: task.title ?? "",
      description: task.description ?? "",
      createdAt: task.created_at ?? null,
      completedAt: task.completed_at ?? null,
      completed: task.completed_at != null,
      deleted: Boolean(task.is_deleted ?? task.deleted ?? task.deleted_at),
      raw: task
    };
    if (normalized.id === undefined || normalized.id === null) {
      console.error("Task list item is missing its id:", task);
    }
    return normalized;
  }
  function currentTasks() { return cache[currentType] || []; }
  function findTask(id) {
    for (const type of types) {
      const found = (cache[type] || []).find(task => String(task.id) === String(id));
      if (found) return found;
    }
    return null;
  }
  function formatDate(value) {
    if (!value) return "—";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return String(value);
    return new Intl.DateTimeFormat("en-US", { dateStyle: "medium", timeStyle: "short" }).format(date);
  }
  function showListMessage(text, state) {
    const $msg = $("#task-list-message").removeClass("is-error is-success is-info").text(text || "");
    if (state) $msg.addClass("is-" + state);
  }
  function hasErrorsField(body) {
    return body && Object.prototype.hasOwnProperty.call(body, "errors");
  }
  function errorMessages(value) {
    if (value == null) return [];
    if (Array.isArray(value)) return value.flatMap(errorMessages);
    if (typeof value === "object") {
      return Object.entries(value).flatMap(([key, val]) => errorMessages(val).map(message => key ? key + ": " + message : message));
    }
    return [String(value)];
  }
  function showFailure(body, $container, fieldTargets, fallback) {
    if (hasErrorsField(body)) {
      const errors = body.errors || {};
      if (fieldTargets) {
        Object.keys(fieldTargets).forEach(field => {
          const $target = fieldTargets[field];
          $target.text("");
          const matchingKeys = Object.keys(errors).filter(key => key.toLowerCase() === field.toLowerCase());
          $target.text(matchingKeys.flatMap(key => errorMessages(errors[key])).join("; "));
        });
      }
      const unknownFields = Object.entries(errors).filter(([key]) => !fieldTargets || !Object.keys(fieldTargets).some(field => field.toLowerCase() === key.toLowerCase()));
      const formatted = unknownFields.flatMap(([, value]) => errorMessages(value)).join("\n");
      if ($container) $container.text(formatted);
      return;
    }
    if ($container) $container.text(body.message || fallback || "");
  }
  function iconFor(type) {
    const icons = {
      uncompleted: "<svg viewBox='0 0 24 24'><circle cx='12' cy='12' r='8.5'/><path d='M12 7.5v4.7l3 1.8'/></svg>",
      completed: "<svg viewBox='0 0 24 24'><circle cx='12' cy='12' r='8.5'/><path d='m8.5 12 2.3 2.3 4.8-5'/></svg>",
      deleted: "<svg viewBox='0 0 24 24'><path d='M4.5 7h15M9 7V4.5h6V7m3 0-.8 12H6.8L6 7m4 3v6m4-6v6'/></svg>"
    };
    return icons[type] || icons.uncompleted;
  }
  function renderList() {
    const tasks = currentTasks();
    const $list = $("#task-list").empty();
    types.forEach(type => $("#count-" + type).text(cache[type] ? cache[type].length : ""));
    if (currentType === "deleted" && tasks.length) {
      const allSelected = tasks.length > 0 && tasks.every(task => selectedForBulk.has(String(task.id)));
      const $bulk = $('<div class="trash-bulk-toolbar"></div>')
        .append($('<label class="trash-select-all"></label>').append(
          $('<input type="checkbox" class="trash-select-all-check">').prop("checked", allSelected),
          $('<span>Select all</span>')
        ))
        .append($('<button type="button" class="button button-danger button-small trash-bulk-delete" disabled>Delete selected permanently</button>'));
      $bulk.find('.trash-select-all-check').on('change', function () {
        if (this.checked) tasks.forEach(task => selectedForBulk.add(String(task.id)));
        else tasks.forEach(task => selectedForBulk.delete(String(task.id)));
        renderList();
      });
      $bulk.find('.trash-bulk-delete').prop('disabled', !tasks.some(task => selectedForBulk.has(String(task.id)))).on('click', bulkDeletePermanently);
      $list.append($bulk);
    }

    if (!tasks.length) {
      const emptyCopy = currentType === "deleted" ? "Trash is empty" : currentType === "completed" ? "No completed tasks yet" : "No uncompleted tasks. Great job!";
      $list.append($('<div class="task-empty-state"></div>').append($('<div class="task-empty-mark">✓</div>'), $('<h3></h3>').text(emptyCopy), $('<p></p>').text(currentType === "deleted" ? "Deleted tasks will appear here." : "Create a task to see it here.")));
      animateNextList = false;
      return;
    }
    tasks.forEach(function (task) {
      const selected = String(task.id) === String(selectedTaskId);
      const $row = $("<article class='task-row' tabindex='0' role='button' aria-label='Open task details'></article>").toggleClass("is-selected", selected);
      $row.on("click", function (event) {
        if ($(event.target).closest("button").length) return;
        selectedTaskId = task.id;
        renderList();
        renderDetails(task);
      }).on("keydown", function (event) {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          selectedTaskId = task.id;
          renderList();
          renderDetails(task);
        }
      });
      const $check = $("<button type='button' class='task-check' aria-label='Mark as completed'></button>").toggleClass("is-checked", task.completed).attr("aria-pressed", String(task.completed)).prop("disabled", currentType === "deleted").html(task.completed ? "✓" : "");
      $check.on("click", function (event) { event.stopPropagation(); if (currentType !== "deleted") toggleTask(task); });
      const $body = $("<button type='button' class='task-row-body'></button>").append($('<span class="task-row-title"></span>').text(task.title).attr("title", task.title), $('<span class="task-row-description"></span>').text(task.description || ""));
      $body.on("click", function () { selectedTaskId = task.id; renderList(); renderDetails(task); });
      const $actions = $("<div class='task-row-actions'></div>");
      if (currentType === "deleted") {
        const $restore = $("<button type='button' class='task-action' title='Restore' aria-label='Restore'>↶</button>").on("click", function (event) { event.stopPropagation(); restoreTask(task); });
        const $permanent = $('<button type="button" class="task-action task-permanent-delete" title="Delete permanently" aria-label="Delete permanently"></button>').html("<svg viewBox='0 0 24 24' aria-hidden='true'><path d='M4.5 7h15M9 7V4.5h6V7m3 0-.8 12H6.8L6 7m4 3v6m4-6v6'/></svg>").on("click", function (event) { event.stopPropagation(); deletePermanently(task); });
        $actions.append($restore, $permanent);
      } else {
        const $delete = $("<button type='button' class='task-action' title='Move to trash' aria-label='Move to trash'></button>").html("<svg viewBox='0 0 24 24' aria-hidden='true'><path d='M4.5 7h15M9 7V4.5h6V7m3 0-.8 12H6.8L6 7m4 3v6m4-6v6'/></svg>").on("click", function (event) { event.stopPropagation(); deleteTask(task); });
        $actions.append($delete);
      }
      if (currentType === "deleted") {
        const $select = $('<input type="checkbox" class="task-select-check" aria-label="Select task for permanent deletion">').prop("checked", selectedForBulk.has(String(task.id)));
        $select.on("click", function (event) { event.stopPropagation(); }).on("change", function () {
          if (this.checked) selectedForBulk.add(String(task.id)); else selectedForBulk.delete(String(task.id));
          renderList();
        });
        $row.append($select);
      }
      $row.append($check, $body, $actions);
      $list.append($row);
    });
    if (animateNextList) {
      $list.addClass("animate-tab-change");
      const $lastRow = $list.find(".task-row").last();
      $lastRow.one("animationend", function () { $list.removeClass("animate-tab-change"); });
    }
    animateNextList = false;
  }
  function renderDetails(task) {
    if (!task) {
      $("#task-detail-empty").prop("hidden", false);
      $("#task-detail-content").prop("hidden", true).empty();
      return;
    }
    const $detail = $("#task-detail-content").empty().prop("hidden", false);
    $("#task-detail-empty").prop("hidden", true);
    const $head = $("<div class='detail-header'></div>");
    const $check = $("<button type='button' class='task-check detail-check' aria-label='Toggle completion'></button>").toggleClass("is-checked", task.completed).attr("aria-pressed", String(task.completed)).prop("disabled", currentType === "deleted").text(task.completed ? "✓" : "");
    $check.on("click", function () { if (currentType !== "deleted") toggleTask(task); });
    const $title = $("<h2 class='detail-title'></h2>").text(task.title);
    $head.append($check, $title);
    const $description = $("<div class='detail-description'></div>").text(task.description || "No description provided.");
    const $dates = $("<div class='detail-dates'></div>");
    $dates.append($('<div class="detail-date-card"></div>').append($('<span class="detail-date-label"></span>').text("Created"), $('<span class="detail-date-value"></span>').text(formatDate(task.createdAt))));
    $dates.append($('<div class="detail-date-card"></div>').append($('<span class="detail-date-label"></span>').text("Completed"), $('<span class="detail-date-value"></span>').text(formatDate(task.completedAt))));
    const $actions = $("<div class='detail-actions'></div>");
    if (currentType === "deleted") {
      $actions.append($('<button type="button" class="button button-secondary button-small"></button>').text("Restore task").on("click", function () { restoreTask(task); }));
      $actions.append($('<button type="button" class="button button-danger button-small"></button>').text("Delete permanently").on("click", function () { deletePermanently(task); }));
    } else {
      $actions.append($('<button type="button" class="button button-secondary button-small detail-edit"></button>').text("Edit task").on("click", function () { renderEditForm(task); }));
      $actions.append($('<button type="button" class="button button-ghost button-small detail-delete"></button>').text("Delete task").on("click", function () { deleteTask(task); }));
    }
    $detail.append($head, $description, $dates, $actions);
  }
  function renderEditForm(task) {
    const $detail = $("#task-detail-content").empty().prop("hidden", false);
    const $form = $("<form class='task-edit-form'></form>");
    const $title = $("<input name='title' maxlength='255' required />").val(task.title);
    const $description = $("<textarea name='description' rows='5'></textarea>").val(task.description || "");
    const $titleError = $('<span class="field-error task-field-error"></span>');
    const $descriptionError = $('<span class="field-error task-field-error"></span>');
    $form.append($('<label>Title</label>').append($title, $titleError), $('<label>Description</label>').append($description, $descriptionError));
    $form.append($('<p class="task-inline-message edit-task-message" role="alert"></p>'));
    const $actions = $('<div class="detail-actions"></div>')
      .append($('<button type="button" class="button button-secondary button-small">Cancel</button>').on("click", function () { renderDetails(task); }))
      .append($('<button type="submit" class="button button-primary button-small edit-task-save">Save changes</button>'));
    $form.append($actions);
    $form.on("submit", function (event) {
      event.preventDefault();
      const payload = { title: $title.val().trim(), description: $description.val().trim() };
      if (!payload.title) return;
      const $save = $form.find('.edit-task-save').prop('disabled', true);
      window.TaskTrackerAPI.tasks.update(task.id, payload).done(function (response) {
        const updated = Object.assign({}, task, payload);
        const body = parseResponse(response);
        const source = body.task || body.data?.task || body.data;
        if (source && typeof source === 'object' && source.id !== undefined) Object.assign(updated, normalize(source));
        types.forEach(t => { if (cache[t]) cache[t] = cache[t].map(item => String(item.id) === String(task.id) ? updated : item); });
        selectedTaskId = updated.id;
        renderList(); renderDetails(updated);
      }).fail(function (xhr) {
        const body = xhr.responseJSON || parseResponse(xhr.responseText);
        showFailure(body, $form.find('.edit-task-message'), { title: $titleError, description: $descriptionError }, 'Unable to update task. Please try again.');
      }).always(function () { $save.prop('disabled', false); });
    });
    $detail.append($form);
    $title.trigger('focus');
  }
  function setLoadingList(loading) {
    isLoading = loading;
    $("#task-list").toggleClass("is-loading", loading);
  }
  function loadTasks(type, force) {
    if (cache[type] && !force) {
      renderList();
      renderDetails(findTask(selectedTaskId));
      return $.Deferred().resolve(cache[type]).promise();
    }
    setLoadingList(true);
    showListMessage("");
    return window.TaskTrackerAPI.tasks.list(type).done(function (response) {
      cache[type] = unwrapList(parseResponse(response)).map(normalize);
      renderList();
      renderDetails(findTask(selectedTaskId));
    }).fail(function (xhr) {
      const body = xhr.responseJSON || parseResponse(xhr.responseText);
      if (hasErrorsField(body)) showFailure(body, $("#task-list-message"));
      else showListMessage(body.message || "", "error");
    }).always(function () { setLoadingList(false); });
  }
  function reloadAfterMutation(type, taskId) {
    types.forEach(t => { cache[t] = null; });
    selectedTaskId = taskId ?? null;
    loadTasks(type, true);
  }
  function toggleTask(task) {
    if (task?.id === undefined || task?.id === null) { showListMessage("This task has no valid ID.", "error"); return; }
    showListMessage("");
    window.TaskTrackerAPI.tasks.toggleCompletion(task.id).done(function () {
      reloadAfterMutation(currentType, null);
    }).fail(function (xhr) {
      const body = xhr.responseJSON || parseResponse(xhr.responseText);
      if (hasErrorsField(body)) showFailure(body, $("#task-list-message"));
      else showListMessage(body.message || "", "error");
    });
  }
  function deleteTask(task) {
    if (task?.id === undefined || task?.id === null) { showListMessage("This task has no valid ID.", "error"); return; }
    showListMessage("");
    window.TaskTrackerAPI.tasks.remove(task.id).done(function () {
      reloadAfterMutation(currentType, null);
    }).fail(function (xhr) {
      const body = xhr.responseJSON || parseResponse(xhr.responseText);
      if (hasErrorsField(body)) showFailure(body, $("#task-list-message"));
      else showListMessage(body.message || "", "error");
    });
  }
  function deletePermanently(task) {
    if (task?.id === undefined || task?.id === null) { showListMessage("This task has no valid ID.", "error"); return; }
    if (!window.confirm("Permanently delete this task? This cannot be undone.")) return;
    showListMessage("");
    window.TaskTrackerAPI.tasks.removeForce(task.id).done(function () {
      selectedForBulk.delete(String(task.id));
      reloadAfterMutation(currentType, null);
    }).fail(function (xhr) {
      const body = xhr.responseJSON || parseResponse(xhr.responseText);
      if (hasErrorsField(body)) showFailure(body, $("#task-list-message"));
      else showListMessage(body.message || "Unable to permanently delete task.", "error");
    });
  }
  function bulkDeletePermanently() {
    const ids = currentTasks().filter(task => selectedForBulk.has(String(task.id))).map(task => task.id);
    if (!ids.length) return;
    if (!window.confirm("Permanently delete " + ids.length + " selected task(s)? This cannot be undone.")) return;
    showListMessage("");
    window.TaskTrackerAPI.tasks.removeBulk(ids).done(function () {
      ids.forEach(id => selectedForBulk.delete(String(id)));
      reloadAfterMutation(currentType, null);
    }).fail(function (xhr) {
      const body = xhr.responseJSON || parseResponse(xhr.responseText);
      if (hasErrorsField(body)) showFailure(body, $("#task-list-message"));
      else showListMessage(body.message || "Unable to permanently delete selected tasks.", "error");
    });
  }
  function restoreTask(task) {
    if (task?.id === undefined || task?.id === null) { showListMessage("This task has no valid ID.", "error"); return; }
    showListMessage("");
    window.TaskTrackerAPI.tasks.restore(task.id).done(function () {
      reloadAfterMutation(currentType, null);
    }).fail(function (xhr) {
      const body = xhr.responseJSON || parseResponse(xhr.responseText);
      if (hasErrorsField(body)) showFailure(body, $("#task-list-message"));
      else showListMessage(body.message || "", "error");
    });
  }
  function switchSection(type) {
    if (!types.includes(type)) return;
    currentType = type;
    selectedTaskId = null;
    selectedForBulk.clear();
    animateNextList = true;
    $("#add-task-button").toggle(type === "uncompleted");
    if (type !== "uncompleted") $("#quick-create-form").prop("hidden", true);
    $(".task-nav-item").removeClass("is-active").filter('[data-task-type="' + type + '"]').addClass("is-active");
    $("#section-title").text(titles[type]);
    $("#section-subtitle").text(subtitles[type]);
    renderDetails(null);
    loadTasks(type, false);
  }
  function bindCreateTask() {
    const $form = $("#quick-create-form");
    $("#add-task-button").on("click", function () {
      $form.prop("hidden", !$form.prop("hidden"));
      if (!$form.prop("hidden")) $form.find('[name="title"]').trigger("focus");
    });
    $("#cancel-create").on("click", function () { $form[0].reset(); $form.prop("hidden", true); $("#create-task-message").text(""); });
    $form.on("submit", function (event) {
      event.preventDefault();
      const payload = { title: $form.find('[name="title"]').val().trim(), description: $form.find('[name="description"]').val().trim() };
      if (!payload.title) return;
      const $submit = $("#create-task-submit").prop("disabled", true);
      const $titleError = $form.find(".create-title-error");
      const $descriptionError = $form.find(".create-description-error");
      $titleError.text(""); $descriptionError.text("");
      $("#create-task-message").text("");
      window.TaskTrackerAPI.tasks.create(payload).done(function () {
        $form[0].reset(); $form.prop("hidden", true);
        reloadAfterMutation(currentType, null);
      }).fail(function (xhr) {
        const body = xhr.responseJSON || parseResponse(xhr.responseText);
        showFailure(body, $("#create-task-message"), { title: $titleError, description: $descriptionError });
      }).always(function () { $submit.prop("disabled", false); });
    });
  }
  $(function () {
    $(".task-nav-item").on("click", function () { switchSection($(this).data("task-type")); });
    bindCreateTask();
    const session = window.TaskTrackerSession || window.TaskTrackerAPI.user();
    session.done(function (raw) {
      const response = parseResponse(raw);
      const user = response.user || response.data?.user || response.data || response;
      if (!user || !user.id && !user.name && !user.email) { window.location.replace("/login/"); return; }
      document.body.classList.remove("auth-pending");
      document.body.classList.add("is-authenticated");
      loadTasks(currentType, false);
    }).fail(function (xhr) {
      window.location.replace("/login/");
    });
  });
})(jQuery, window, document);
