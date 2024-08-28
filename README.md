
# Hermes

**Hermes** 是你的郵件小助手，提供了一個帶有 TUI 介面的 CLI SMTP 客戶端。無論是簡單的 CLI 命令還是互動式 TUI 介面，它都能幫助你輕鬆發送電子郵件，讓你瞬間成為電子郵件界的信使之神！👈 AI 真會唬爛

---

## 功能特點

- **CLI 模式**：快速發送電子郵件，無需圖形介面。
- **TUI 模式**：提供互動式的文字用戶界面，讓你更直觀地操作發送郵件流程。
- **多選項配置**：支援從命令行傳遞發件人、收件人、主題等詳細信息。
- **輕鬆發送**：可配置 SMTP 主機與端口，支援不同郵件伺服器。

---

## 安裝

在本地安裝 Hermes 並開始使用：

*...待完成*

---

## 使用說明

### CLI 模式

傳統的命令行執行方式，適合喜歡使用命令行發送郵件的用戶。

```bash
hermes directSendMail [flags]
```

#### 可用參數

| 參數             | 描述                                           |
|------------------|------------------------------------------------|
| `--contents`     | 設定郵件內容                                   |
| `--from`         | 設定發件人電子郵件地址（如：`someone@example.com`） |
| `--host`         | 設定 MTA 主機名稱（如：`smtp.gmail.com`）        |
| `--port`         | 設定 SMTP 伺服器端口（如：`25`）               |
| `--subject`      | 設定郵件主題                                   |
| `--to`           | 設定收件人電子郵件地址（如：`someone@example.com`） |
| `-h`, `--help`   | 查看幫助                                       |

#### 範例

快速發送郵件：

```bash
hermes directSendMail --from="you@example.com" --to="friend@example.com" --subject="Hello" --contents=Hello from Hermes!" --host=smtp.gmail.com" --port="587"
```

---

### Burst 模式

爆發模式發送郵件，一次併發大量郵件發送模式。

```bash
hermes burst [flags]
```

#### 範例

爆發模式發送郵件：

```bash
hermes burst --quantity="1000" --host=smtp.gmail.com" --port="587"
```

---

#### 可用參數

| 參數               | 描述                                         |
|--------------------|----------------------------------------------|
| `--host`           | MTA 主機名稱（例如：`smtp.gmail.com`）         |
| `--port`           | 端口號（例如：`25`）                          |
| `--quantity`       | 要發送的郵件數量                              |
| `-h`, `--help`     | 查看幫助                                     |

---

### TUI 模式

啟動互動式文字用戶界面，適合需要介面操作發送郵件的用戶。

```bash
hermes start-tui
# 或
hermes start-tui [flags]
```

#### 可用參數

| 參數             | 描述     |
|------------------|----------|
| `-h`, `--help`   | 查看幫助 |

---

## 示例

![Demo](./assets/imgs/hermes.gif)

---
