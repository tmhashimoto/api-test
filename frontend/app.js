// ブラウザからはホスト側に公開した API ポートへ向ける
const BROWSER_API = "http://localhost:8080";

document.getElementById("api-base-label").textContent = BROWSER_API;

function pretty(value) {
  return typeof value === "string" ? value : JSON.stringify(value, null, 2);
}

async function readError(res) {
  const text = await res.text();
  try {
    return JSON.parse(text);
  } catch {
    return text || res.statusText;
  }
}

function renderExplain(container, data, mode) {
  const diffs = (data.differences || [])
    .map((d, i) => `${i + 1}. ${d}`)
    .join("\n");

  const beforeLabel =
    mode === "marshal"
      ? "① Before（Goの中）— 構造体。型がある。Goしか読めない"
      : "① Before（ネットワーク上）— JSON。ただのテキスト";
  const afterLabel =
    mode === "marshal"
      ? "② After（ネットワーク上）— JSON。ただのテキスト。誰でも読める"
      : "② After（Goの中）— 構造体。型がある。Goが扱える形";

  const beforeBody = mode === "marshal" ? data.go_code : data.json_text;
  const afterBody = mode === "marshal" ? data.json_text : data.go_code;

  container.innerHTML = `
    <p class="explain-desc">${escapeHtml(data.description)}</p>
    <div class="compare">
      <div class="compare-pane">
        <h4>${escapeHtml(beforeLabel)}</h4>
        <pre>${escapeHtml(beforeBody)}</pre>
      </div>
      <div class="compare-arrow" aria-hidden="true">${mode === "marshal" ? "Marshal →" : "Unmarshal →"}</div>
      <div class="compare-pane">
        <h4>${escapeHtml(afterLabel)}</h4>
        <pre>${escapeHtml(afterBody)}</pre>
      </div>
    </div>
    <h4 class="explain-sub">見た目は似ているが、違う点</h4>
    <pre class="explain-diffs">${escapeHtml(diffs)}</pre>
    <p class="explain-msg">${escapeHtml(data.message)}</p>
  `;
}

document.getElementById("btn-marshal").addEventListener("click", async () => {
  const out = document.getElementById("marshal-out");
  out.textContent = "取得中...";
  try {
    const res = await fetch(`${BROWSER_API}/api/explain/marshal`);
    if (!res.ok) throw new Error(pretty(await readError(res)));
    const data = await res.json();
    renderExplain(out, data, "marshal");
  } catch (err) {
    out.textContent = `エラー: ${err.message}\n\nAPI (${BROWSER_API}) が起動しているか確認してください。`;
  }
});

document.getElementById("btn-unmarshal").addEventListener("click", async () => {
  const out = document.getElementById("unmarshal-out");
  const jsonText = document.getElementById("json-input").value.trim();
  out.textContent = "変換中...";
  try {
    const res = await fetch(`${BROWSER_API}/api/explain/unmarshal`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ json_text: jsonText }),
    });
    if (!res.ok) throw new Error(pretty(await readError(res)));
    const data = await res.json();
    renderExplain(out, data, "unmarshal");
  } catch (err) {
    out.textContent = `エラー: ${err.message}`;
  }
});

async function loadBooks() {
  const list = document.getElementById("book-list");
  const log = document.getElementById("api-log");
  list.innerHTML = "";
  log.textContent = "GET /api/books ...";

  try {
    const res = await fetch(`${BROWSER_API}/api/books`);
    if (!res.ok) throw new Error(pretty(await readError(res)));
    const books = await res.json();

    if (!books.length) {
      list.innerHTML = "<li>まだ書籍がありません</li>";
    } else {
      for (const book of books) {
        const li = document.createElement("li");
        li.innerHTML = `
          <div>
            <strong>${escapeHtml(book.title)}</strong>
            <div class="meta">${escapeHtml(book.author)} / ${book.year}</div>
          </div>
          <div class="meta">id: ${book.id}</div>
        `;
        list.appendChild(li);
      }
    }

    log.textContent = [
      "【サーバー側】books（構造体のスライス）を json.Marshal → HTTPレスポンス",
      "【ブラウザ側】受け取った JSON テキストを JavaScript のオブジェクトに変換",
      "",
      pretty(books),
    ].join("\n");
  } catch (err) {
    log.textContent = `エラー: ${err.message}`;
  }
}

document.getElementById("btn-reload").addEventListener("click", loadBooks);

document.getElementById("book-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.currentTarget;
  const log = document.getElementById("api-log");
  const payload = {
    title: form.title.value.trim(),
    author: form.author.value.trim(),
    year: Number(form.year.value),
  };

  log.textContent = [
    "送信する JSON:",
    pretty(payload),
    "",
    "POST /api/books ...",
  ].join("\n");

  try {
    const res = await fetch(`${BROWSER_API}/api/books`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) throw new Error(pretty(await readError(res)));
    const created = await res.json();
    log.textContent = [
      "【ブラウザ】フォームの値 → JSON.stringify（ほぼ Marshal と同じ役割）",
      "【サーバー】json.Decoder.Decode = Unmarshal → Book 構造体へ",
      "【サーバー】保存後、再び Marshal してレスポンス返却",
      "",
      "作成された書籍:",
      pretty(created),
    ].join("\n");
    form.reset();
    form.year.value = "2026";
    await loadBooks();
  } catch (err) {
    log.textContent = `エラー: ${err.message}`;
  }
});

function escapeHtml(str) {
  return String(str)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

loadBooks();
