# Phase 19: Blog Settings Enhancement - UI Design

**Created:** 2026-05-09
**Phase:** 19 (Blog Settings Enhancement)
**Requirements:** SETT-01~05

## Summary

在设置页面的博客卡片中显示 Blog URL 和 Feed URL，并支持在统一编辑表单内修改这两个字段，提交前进行 URL 格式验证（HTTP/HTTPS）。

## UI Design Decisions

### D-01: URL 显示位置 — 两列并排

**决策：** Blog URL 和 Feed URL 左右并列显示在卡片中。

**Why:**
- 信息一目了然，两 URL 并列对比清晰
- 与现有卡片布局兼容（已有名称、文章数等信息）

**How to apply:**
```html
<div style="display: flex; gap: 24px;">
  <div>
    <div class="label">Blog URL</div>
    <a href="{{.Blog.URL}}" target="_blank">{{.Blog.URL}}</a>
  </div>
  <div>
    <div class="label">Feed URL</div>
    <a href="{{.Blog.FeedURL}}" target="_blank">{{.Blog.FeedURL}}</a>
  </div>
</div>
```

---

### D-02: 编辑交互模式 — 统一编辑表单

**决策：** 点击 Edit 进入统一表单，编辑 name + Blog URL + Feed URL + category。

**Why:**
- 与现有 Edit 流程一致，用户习惯已建立
- 一次编辑所有字段，无需多次交互
- 实现简单，复用现有 HTMX 模式

**How to apply:**
扩展 `blog-edit-form.gohtml`，在现有 name 和 category 之间添加两个输入框：

```html
<form hx-put="/blogs/{{.Blog.ID}}" hx-target="#blog-{{.Blog.ID}}" hx-swap="outerHTML">
  <!-- 名称 -->
  <input type="text" name="name" value="{{.Blog.Name}}" required>

  <!-- Blog URL -->
  <div class="blog-edit-url">
    <label class="label">Blog URL</label>
    <input type="url" name="url" value="{{.Blog.URL}}" pattern="^https?://.*">
  </div>

  <!-- Feed URL -->
  <div class="blog-edit-feed-url">
    <label class="label">Feed URL</label>
    <input type="url" name="feed_url" value="{{.Blog.FeedURL}}" pattern="^https?://.*">
  </div>

  <!-- 分类 -->
  <select name="category_id">...</select>

  <!-- Save / Cancel -->
  <button type="submit">Save</button>
  <button type="button" hx-get="/blogs/{{.Blog.ID}}">Cancel</button>
</form>
```

---

### D-03: 验证反馈方式 — Inline 错误提示

**决策：** URL 验证失败时，在输入框下方显示红色错误文字，阻止提交。

**Why:**
- 错误位置精确，用户立即知道哪个字段有问题
- 阻止提交无效数据，减少无效请求
- 与表单验证通用风格一致

**How to apply:**

**前端验证（HTML5 pattern）：**
```html
<input type="url" name="url" pattern="^https?://.*" required>
```

**错误提示样式：**
```css
.blog-edit-url input:invalid,
.blog-edit-feed-url input:invalid {
  border-color: var(--danger);
}

.blog-edit-url input:invalid + .error-message,
.blog-edit-feed-url input:invalid + .error-message {
  display: block;
  color: var(--danger);
  font-size: 13px;
  margin-top: 4px;
}
```

**错误消息内容：**
- `URL 必须以 http:// 或 https:// 开头`

---

## Display Layout

### 非编辑状态

```
┌─────────────────────────────────────────────────┐
│ 博客名称                                         │
│ Blog URL           │ Feed URL                    │
│ https://blog.com   │ https://blog.com/feed.xml   │
│ 42 articles · 技术分类                           │
│                    │ [Edit] [Remove]             │
└─────────────────────────────────────────────────┘
```

### 编辑状态

```
┌─────────────────────────────────────────────────┐
│ 名称:   [博客名称_____________________]          │
│ Blog URL: [https://blog.com________]            │
│ Feed URL: [https://blog.com/feed.xml]           │
│ 分类:   [技术 ▼]                                 │
│                                                  │
│ [Save] [Cancel]                                  │
└─────────────────────────────────────────────────┘
```

---

## Data Flow

1. **点击 Edit** → `GET /blogs/{id}/edit` → 返回扩展编辑表单
2. **提交 Save** → `PUT /blogs/{id}` → 后端验证 + 更新数据库
3. **成功** → 返回 `blog-display-row.gohtml` 刷新卡片
4. **失败** → 返回编辑表单 + 错误提示（后端兜底验证）

---

## Component Changes

| 文件 | 变更 |
|------|------|
| `blog-display-row.gohtml` | 添加两列 URL 显示（Blog URL + Feed URL） |
| `blog-edit-form.gohtml` | 添加 Blog URL / Feed URL 输入框 + 错误提示位置 |
| `internal/server/handlers.go` | `handleBlogEdit` 传递完整 Blog 数据；`handleBlogUpdate` 解析 + 验证 URL 参数 |
| `internal/storage/database.go` | 添加 `UpdateBlog(blogID, name, url, feedURL, categoryID)` 方法 |

---

## Technical Notes

**URL 验证规则：**
- 必须以 `http://` 或 `https://` 开头
- 空值允许（字段 nullable）
- 验证在提交前进行（前端 HTML5 pattern + 后端兜底）

**后端验证实现：**
```go
func validateURL(url string) error {
    if url == "" {
        return nil // nullable
    }
    if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
        return fmt.Errorf("URL 必须以 http:// 或 https:// 开头")
    }
    return nil
}
```

---

## Out of Scope

- Blog URL 变化后自动更新文章 URL（需手动 rescan）
- Feed URL 变化后自动重新扫描（需手动触发）
- 编辑时的预览功能（Phase 20 内容）

---

*Design validated: 2026-05-09*
*Phase: 19-Blog Settings Enhancement*