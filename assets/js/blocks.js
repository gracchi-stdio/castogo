import Sortable from "sortablejs";

let sortableInstance = null;

export function initBlockSorter() {
  if (sortableInstance) {
    sortableInstance.destroy();
    sortableInstance = null;
  }

  const list = document.getElementById("block-list");
  if (!list) return;

  sortableInstance = Sortable.create(list, {
    animation: 150,
    handle: "[data-block-id]",
    forceFallback: true,
    fallbackClass: "sortable-fallback",
    ghostClass: "sortable-ghost",
    chosenClass: "sortable-chosen",
    onStart: function () {
      document.body.classList.add("is-dragging");
    },
    onEnd: function () {
      document.body.classList.remove("is-dragging");

      const items = list.querySelectorAll("[data-block-id]");
      const blockIds = Array.from(items).map((el) =>
        parseInt(el.getAttribute("data-block-id"), 10)
      );

      const match = window.location.pathname.match(/\/admin\/pages\/(\d+)/);
      if (!match) return;
      const pageID = match[1];

      fetch(`/admin/pages/${pageID}/blocks/reorder`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ block_ids: blockIds }),
      });
    },
  });
}
