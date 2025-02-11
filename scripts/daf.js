function submitDafForm() {
  // 取得表單資料+整理資料
  const form = document.getElementById("dafFilterForm");
  const formData = new FormData(form);
  formData.append("store", "daf");
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
    text: "請稍候",
    allowOutsideClick: false,
    didOpen: () => {
      Swal.showLoading();
    },
  });

  fetch(`${url}?${params}`)
    .then((response) => response.json())
    .then((data) => {
      Swal.close(); // 關閉讀取中的遮罩
      document.getElementById("shopname").innerText = "D+AF";
      const tableBody = document.querySelector("tbody");
      tableBody.innerHTML = "";

      // 沒有找到符合條件的結果
      if (data ==null ||data.length == 0) {
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
                        <td><a href="${shoe.url}" target="_blank">連結</a></td>
                        <td>${highlightSizes(shoe.size, selectedSizeText)}</td>
                        <td>${shoe.color.join(", ")}</td>
                        <td>D+AF</td>
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

