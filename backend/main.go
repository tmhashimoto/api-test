package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Book は Go の構造体です。
// メモリ上では「型付きの値」として扱われますが、
// ネットワーク越しにはそのまま送れません。
// json タグは「JSON に変換するときのキー名」を指定します。
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

// ExplainResponse は学習用のレスポンスです。
// 「何が起きたか」をフロントエンドで見えるようにします。
type ExplainResponse struct {
	Step        string   `json:"step"`
	Description string   `json:"description"`
	GoCode      string   `json:"go_code"`      // Goのソースに近い見た目（構造体）
	JSONText    string   `json:"json_text"`    // ネットワーク上のテキスト（JSON）
	Differences []string `json:"differences"`  // 見た目が似ていても違う点
	Message     string   `json:"message"`
}

// bookAsGoCode は「構造体は Go の書き方」だと分かるように整形します。
// JSON と並べると、同じデータでも表現が違うことが見えます。
func bookAsGoCode(b Book) string {
	return fmt.Sprintf(
		"Book{\n"+
			"    ID:     %d,           // 型: int\n"+
			"    Title:  %q, // 型: string\n"+
			"    Author: %q,     // 型: string\n"+
			"    Year:   %d,         // 型: int\n"+
			"}",
		b.ID, b.Title, b.Author, b.Year,
	)
}

func explainDifferences() []string {
	return []string{
		"構造体のフィールド名は ID / Title（大文字）。JSON では id / title（小文字のキー）",
		"構造体には「int」「string」という型がある。JSON はただの文字なので型情報はない",
		"構造体は Go のメモリ上の値。JSON は HTTP で送れるテキストの共通フォーマット",
		"Python やブラウザは Go の構造体を読めないが、JSON なら読める",
	}
}

var (
	mu    sync.Mutex
	books = []Book{
		{ID: 1, Title: "Go言語で学ぶAPI", Author: "山田太郎", Year: 2024},
		{ID: 2, Title: "JSONと仲良くなる本", Author: "佐藤花子", Year: 2023},
	}
	nextID = 3
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// handleListBooks: 構造体のスライスを JSON に「マーシャル」して返す
// Webアプリでの典型例: 一覧画面用の GET API
func handleListBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// ★ Marshal: Goの構造体 → JSONのバイト列
	// ブラウザや他の言語は「Goの構造体」を直接理解できないので、
	// 共通フォーマットである JSON に変換してから送ります。
	data, err := json.Marshal(books)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// handleCreateBook: リクエストの JSON を構造体に「アンマーシャル」して保存
// Webアプリでの典型例: フォーム送信・新規作成の POST API
func handleCreateBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var book Book

	// ★ Unmarshal: JSONのバイト列 → Goの構造体
	// ブラウザから届くのは「文字列の JSON」です。
	// Go で安全に扱うには、まず構造体に変換します。
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "JSONの形式が正しくありません: "+err.Error(), http.StatusBadRequest)
		return
	}

	if book.Title == "" || book.Author == "" {
		http.Error(w, "title と author は必須です", http.StatusBadRequest)
		return
	}

	mu.Lock()
	book.ID = nextID
	nextID++
	books = append(books, book)
	mu.Unlock()

	// 保存した構造体を、また JSON にマーシャルして返します
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(book)
}

// handleExplainMarshal: マーシャルの流れを可視化する学習用エンドポイント
func handleExplainMarshal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	sample := Book{
		ID:     99,
		Title:  "はじめてのJSON",
		Author: "学習くん",
		Year:   2026,
	}

	// 実際に Marshal して、変換後の文字列も一緒に返す
	raw, err := json.MarshalIndent(sample, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ExplainResponse{
		Step:        "marshal",
		Description: "左（Go構造体）→ 右（JSON文字列）へ変換しています。見た目は似ていますが、中身の意味が違います",
		GoCode:      bookAsGoCode(sample),
		JSONText:    string(raw),
		Differences: explainDifferences(),
		Message: "だから API では「構造体のまま」は送れず、必ず JSON などの共通テキストへ Marshal します。",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

// handleExplainUnmarshal: アンマーシャルの流れを可視化する学習用エンドポイント
func handleExplainUnmarshal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	// 生の JSON 文字列をまず読む
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "JSONの形式が正しくありません: "+err.Error(), http.StatusBadRequest)
		return
	}

	jsonBytes, ok := raw["json_text"]
	if !ok {
		http.Error(w, `{"json_text": "{...}"} の形で送ってください`, http.StatusBadRequest)
		return
	}

	// json_text 自体が JSON 文字列なので、一度 string にアンマーシャル
	var jsonText string
	if err := json.Unmarshal(jsonBytes, &jsonText); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var book Book
	if err := json.Unmarshal([]byte(jsonText), &book); err != nil {
		http.Error(w, "アンマーシャル失敗: "+err.Error(), http.StatusBadRequest)
		return
	}

	res := ExplainResponse{
		Step:        "unmarshal",
		Description: "左（JSON文字列）→ 右（Go構造体）へ変換しています。テキストを、型付きのデータに戻しています",
		GoCode:      bookAsGoCode(book),
		JSONText:    jsonText,
		Differences: explainDifferences(),
		Message: "JSON のままでは year が「数字の文字」か「本当の数値」か分かりにくいです。" +
			"構造体へ Unmarshal すると Year は int として扱えます。",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/books", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodOptions:
			handleListBooks(w, r)
		case http.MethodPost:
			handleCreateBook(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/explain/marshal", withCORS(handleExplainMarshal))
	mux.HandleFunc("/api/explain/unmarshal", withCORS(handleExplainUnmarshal))

	addr := ":8080"
	log.Printf("API server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
