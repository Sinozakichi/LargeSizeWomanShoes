// 從後端獲取鞋子資料
const url = "http://localhost:8080/filter";

document.addEventListener("DOMContentLoaded", function () {
  const navLinks = document.querySelectorAll(".nav-link");
  const tabPanes = document.querySelectorAll(".tab-pane");

  navLinks.forEach((link) => {
    link.addEventListener("click", function (event) {
      event.preventDefault();

      // 移除所有 nav-link 的 active 類
      navLinks.forEach((nav) => nav.classList.remove("active"));

      // 為當前點擊的 nav-link 添加 active 類
      this.classList.add("active");

      // 移除所有 tab-pane 的 show 和 active 類
      tabPanes.forEach((pane) => pane.classList.remove("show", "active"));

      // 為對應的 tab-pane 添加 show 和 active 類
      const targetPane = document.querySelector(this.getAttribute("href"));
      targetPane.classList.add("show", "active");
    });
  });
});
