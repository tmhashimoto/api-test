# Go構造体とJSONを学ぶ（Docker環境）

初学者向けに、**なぜ API では構造体をそのまま送らず JSON に変換するのか**、
**Marshal / Unmarshal は実際の Web アプリでいつ使うのか** を、動かしながら説明するための小さな環境です。

## 起動方法

```bash
docker compose up --build
# docker 権限がない場合
# sudo docker compose up --build
```

| サービス | URL |
|---------|-----|
| フロントエンド（学習ページ） | http://localhost:3000 |
| Go API | http://localhost:8080 |

停止:

```bash
docker compose down
```

## この環境で学べること

### 1. 構造体と JSON は似ているのに、なぜ変換が必要？

| | Goの構造体 | JSON |
|---|---|---|
| 何か | プログラムのメモリ上の「型付きデータ」 | テキストの共通フォーマット |
| 誰が読める？ | Go だけ | ブラウザ・Python・Java などほぼ全部 |
| ネットワーク向き？ | そのままでは送れない | HTTP のボディとして送れる |

見た目が似ていても、**「Go専用の箱」と「みんなの共通語」は別物**です。
API 通信では、必ず JSON（などの共通形式）に変換します。

### 2. Marshal / Unmarshal はいつ使う？

```text
[ブラウザ] --JSON--> [Go API] --JSON--> [ブラウザ]

送るとき（サーバー→クライアント）: 構造体 → JSON   = Marshal
受け取るとき（クライアント→サーバー）: JSON → 構造体 = Unmarshal
```

実際の Web アプリでの典型例:

- **Marshal**: `GET /api/books` で書籍一覧を返すとき
- **Unmarshal**: `POST /api/books` で新規登録の JSON を受け取るとき

コード上の対応箇所は `backend/main.go` を見てください（コメント付き）。

### 3. 画面の使い方

1. **① Marshal** … 構造体が JSON 文字列になる様子を見る
2. **② Unmarshal** … JSON 文字列を構造体へ戻す様子を見る
3. **③ 書籍メモ** … 本物の API 連携（作成・一覧）を試す

## ディレクトリ構成

```text
api-test/
├── docker-compose.yml
├── backend/          # Go API
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
└── frontend/         # 学習用 UI
    ├── Dockerfile
    ├── nginx.conf
    ├── index.html
    ├── styles.css
    └── app.js
```

## API 一覧

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/books` | 一覧（Marshal の実例） |
| POST | `/api/books` | 作成（Unmarshal の実例） |
| GET | `/api/explain/marshal` | Marshal の可視化 |
| POST | `/api/explain/unmarshal` | Unmarshal の可視化 |
| GET | `/health` | ヘルスチェック |
