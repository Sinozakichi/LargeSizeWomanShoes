function submitAnnsForm() {
  var pass = CheckAnnsFormImformation();
  if (!pass) {
    return;
  }

  // 取得表單資料+整理資料
  const form = document.getElementById("annsFilterForm");
  const formData = new FormData(form);
  formData.append("store", "anns");
  const params = new URLSearchParams(formData).toString();
  const selectedCategoryText = form.querySelector(
    "#searchCat option:checked"
  ).textContent;
  const selectedSizeText = form.querySelector(
    "#searchSize option:checked"
  ).textContent;

  // 顯示讀取中的遮罩
  Swal.fire({
    title: "讀取中...",
    text: "資料量大的時候可能需要一點時間，請稍候",
    allowOutsideClick: false,
    didOpen: () => {
      Swal.showLoading();
    },
  });

  fetch(`${url}?${params}`)
    .then((response) => response.json())
    .then((data) => {
      Swal.close(); // 關閉讀取中的遮罩
      document.getElementById("shopname").innerText = "Ann's";
      const tableBody = document.querySelector("tbody");
      tableBody.innerHTML = "";

      // 沒有找到符合條件的結果
      if (data == null || data.length == 0) {
        Swal.fire({
          icon: "info",
          title: "搜尋結果",
          text: "沒有找到符合條件的結果",
        });
        return;
      }
      data.forEach((shoe) => {
        const row = document.createElement("tr");
        row.innerHTML = `
                          <td>${shoe.name}</td>
                          <td>${shoe.price}</td>
                          <td><img src="${shoe.image}" alt="${
          shoe.name
        }" style="width: 50px; height: auto;"></td>
                          <td><a href="${
                            shoe.url
                          }" target="_blank">連結</a></td>
                          <td>${highlightSizes(
                            shoe.size,
                            selectedSizeText
                          )}</td>
                          <td>${formatShoeColor(shoe)}</td>
                          <td>Ann's</td>
                          <td>${selectedCategoryText}</td>
                      `;
        tableBody.appendChild(row);
      });
      Swal.fire({
        icon: "success",
        title: "資料搜索成功",
        showConfirmButton: false,
        timer: 1500,
      });
    })
    .catch((error) => {
      console.error("Error:", error);
      Swal.fire({
        icon: "error",
        title: "資料搜索失敗",
        text: error.message,
      });
    });
}

function CheckAnnsFormImformation() {
  // 取得表單資料+整理資料
  const form = document.getElementById("annsFilterForm");
  const searchSize = form.querySelector("#searchSize").value;
  const searchCat = form.querySelector("#searchCat").value;

  // 檢查 searchSize 與 searchCat 是否為空值
  if (!searchSize) {
    Swal.fire({
      icon: "warning",
      title: "錯誤",
      text: "請選擇鞋碼尺寸!",
    });
    return false; // 阻止表單提交
  }
  if (!searchCat) {
    Swal.fire({
      icon: "warning",
      title: "錯誤",
      text: "請選擇鞋子類別!",
    });
    return false; // 阻止表單提交
  }
  return true;
}
