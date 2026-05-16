const API_URL = window.COMMENTTREE_API_URL || "http://localhost:8080";

const state = {
  q: "",
  sort: "created_at",
  order: "desc",
  page: 1,
  limit: 10,
  totalPages: 0,
};

const elements = {
  comments: document.querySelector("#comments"),
  template: document.querySelector("#comment-template"),
  status: document.querySelector("#status"),
  searchForm: document.querySelector("#search-form"),
  searchInput: document.querySelector("#search-input"),
  resetSearch: document.querySelector("#reset-search"),
  sortSelect: document.querySelector("#sort-select"),
  orderSelect: document.querySelector("#order-select"),
  limitSelect: document.querySelector("#limit-select"),
  rootForm: document.querySelector("#root-form"),
  prevPage: document.querySelector("#prev-page"),
  nextPage: document.querySelector("#next-page"),
  pageLabel: document.querySelector("#page-label"),
};

elements.searchForm.addEventListener("submit", (event) => {
  event.preventDefault();
  state.q = elements.searchInput.value.trim();
  state.page = 1;
  loadComments();
});

elements.resetSearch.addEventListener("click", () => {
  elements.searchInput.value = "";
  state.q = "";
  state.page = 1;
  loadComments();
});

elements.sortSelect.addEventListener("change", () => {
  state.sort = elements.sortSelect.value;
  state.page = 1;
  loadComments();
});

elements.orderSelect.addEventListener("change", () => {
  state.order = elements.orderSelect.value;
  state.page = 1;
  loadComments();
});

elements.limitSelect.addEventListener("change", () => {
  state.limit = Number(elements.limitSelect.value);
  state.page = 1;
  loadComments();
});

elements.rootForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  await submitComment(new FormData(elements.rootForm));
  elements.rootForm.reset();
});

elements.prevPage.addEventListener("click", () => {
  if (state.page > 1) {
    state.page -= 1;
    loadComments();
  }
});

elements.nextPage.addEventListener("click", () => {
  if (state.page < state.totalPages) {
    state.page += 1;
    loadComments();
  }
});

async function loadComments() {
  setStatus("Загружаем комментарии...");
  elements.comments.innerHTML = "";

  try {
    const params = new URLSearchParams({
      page: String(state.page),
      limit: String(state.limit),
      sort: state.sort,
      order: state.order,
    });

    if (state.q) {
      params.set("q", state.q);
    }

    const response = await fetch(`${API_URL}/comments?${params.toString()}`);
    const page = await readJSON(response);

    if (page.total > 0 && page.items.length === 0 && state.page > 1) {
      state.page -= 1;
      await loadComments();
      return;
    }

    state.totalPages = page.totalPages;
    renderComments(page.items);
    renderPagination(page);
    setStatus(statusText(page));
  } catch (error) {
    renderError(error.message);
    setStatus("Не удалось загрузить комментарии");
  }
}

function renderComments(comments) {
  elements.comments.innerHTML = "";

  if (comments.length === 0) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = state.q ? "Поиск ничего не нашел" : "Комментариев пока нет";
    elements.comments.append(empty);
    return;
  }

  const fragment = document.createDocumentFragment();
  for (const comment of comments) {
    fragment.append(renderComment(comment));
  }
  elements.comments.append(fragment);
}

function renderComment(comment) {
  const node = elements.template.content.firstElementChild.cloneNode(true);

  node.querySelector(".comment-author").textContent = comment.author;
  node.querySelector(".comment-date").textContent = formatDate(comment.createdAt);
  node.querySelector(".comment-date").dateTime = comment.createdAt;
  node.querySelector(".comment-text").textContent = comment.text;

  const replyButton = node.querySelector(".reply-button");
  const deleteButton = node.querySelector(".delete-button");
  const replyForm = node.querySelector(".reply-form");
  const cancelReply = node.querySelector(".cancel-reply");
  const children = node.querySelector(".children");

  replyButton.addEventListener("click", () => {
    replyForm.classList.toggle("hidden");
    if (!replyForm.classList.contains("hidden")) {
      replyForm.querySelector("input").focus();
    }
  });

  cancelReply.addEventListener("click", () => {
    replyForm.reset();
    replyForm.classList.add("hidden");
  });

  replyForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await submitComment(new FormData(replyForm), comment.id);
    replyForm.reset();
    replyForm.classList.add("hidden");
  });

  deleteButton.addEventListener("click", async () => {
    const confirmed = window.confirm("Удалить комментарий и все ответы?");
    if (!confirmed) {
      return;
    }

    await deleteComment(comment.id);
  });

  for (const child of comment.children || []) {
    children.append(renderComment(child));
  }

  return node;
}

async function submitComment(formData, parentId) {
  const payload = {
    author: String(formData.get("author") || "").trim(),
    text: String(formData.get("text") || "").trim(),
  };

  if (parentId) {
    payload.parentId = parentId;
  }

  if (!payload.author || !payload.text) {
    setStatus("Заполните имя и текст комментария");
    return;
  }

  try {
    const response = await fetch(`${API_URL}/comments`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    await readJSON(response);
    setStatus("Комментарий сохранен");
    await loadComments();
  } catch (error) {
    setStatus(error.message);
  }
}

async function deleteComment(id) {
  try {
    const response = await fetch(`${API_URL}/comments/${id}`, { method: "DELETE" });
    await readJSON(response);
    setStatus("Комментарий удален");
    await loadComments();
  } catch (error) {
    setStatus(error.message);
  }
}

async function readJSON(response) {
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload.error || "Ошибка запроса");
  }

  return payload;
}

function renderPagination(page) {
  const totalPages = Math.max(page.totalPages, 1);
  elements.pageLabel.textContent = `${page.page} / ${totalPages}`;
  elements.prevPage.disabled = page.page <= 1;
  elements.nextPage.disabled = page.page >= page.totalPages || page.totalPages === 0;
}

function renderError(message) {
  const error = document.createElement("div");
  error.className = "error-state";
  error.textContent = message;
  elements.comments.replaceChildren(error);
}

function statusText(page) {
  const count = page.total;
  if (count === 0) {
    return "Нет комментариев для отображения";
  }

  return `Найдено корневых веток: ${count}`;
}

function setStatus(message) {
  elements.status.textContent = message;
}

function formatDate(value) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

loadComments();
