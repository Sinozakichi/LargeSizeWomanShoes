//開啟對應鞋店的Fliter
function openFliter(shopname) {
  document.querySelectorAll(".shoearea").forEach(function (element) {
    element.style.display = "none";
  });
  switch (shopname) {
    case "daf":
      document.getElementById("dafArea").style.display = "block";
      break;
    case "anns":
      document.getElementById("annsArea").style.display = "block";
      break;
    case "amai":
      alert("尚未完成");
      break;
    case "gracegift":
      alert("尚未完成");
      break;
  }
}

// 標示出選擇的鞋碼
function highlightSizes(sizes, selectedSize) {
  if (!sizes || sizes.length === 0) {
    return "N/A"; // 如果 sizes 為空值，返回空字符串
  }

  return sizes
    .map((size) => {
      if (size === selectedSize) {
        return `<span style="color: red;">${size}</span>`;
      }
      return size;
    })
    .join(", ");
}

// 將鞋子顏色隔開
function formatShoeColor(shoe) {
  if (!shoe.color) {
    return "N/A"; // 如果 shoe.color 為 null 或 undefined，返回 "N/A"
  }
  return shoe.color.join(", ");
}

// 監聽Tab切換事件
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
