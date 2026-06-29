# 文章 ID 可见与复制 设计文档

日期：2026-06-29

## 背景

BlogWatcher 的 Web UI 浏览文章时，用户无法直接看到文章的数字 ID。当需要把文章 ID 作为 CLI 命令参数（如 `blogwatcher article ...`、收藏/备注相关命令）使用时，必须切换到 CLI 输出才能获得 ID，体验割裂。

## 目标

在文章列表卡片上常驻显示文章 ID，点击即可复制到剪贴板，便于粘贴到 CLI 命令中。

## 非目标

- 不修改后端逻辑、数据模型或 API。
- 不改变 CLI 输出格式。
- 不为 ID 复制做自动化测试（浏览器剪贴板行为难以自动化，采用手动验证）。

## 设计

### 模板：`assets/templates/partials/article-items.gohtml`

在 `.article-meta` 行末尾（`article-time` 之后）追加始终显示的 ID span：

```html
<span class="article-id"
      data-id="{{.ID}}"
      title="点击复制文章 ID（用于 CLI）"
      onclick="copyArticleId(this, event)">#{{.ID}}</span>
```

文章标题链接使用 `stretched-link`，点击卡片任意位置都会打开原文。`copyArticleId` 内部需 `event.stopPropagation()` + `event.preventDefault()` 阻止冒泡与默认行为，避免点了 ID 却跳转原文。

### JS：新增 `assets/static/article.js`，在 `base.gohtml` 引入

全局函数 `copyArticleId(el, event)`：

1. `event.stopPropagation(); event.preventDefault();`
2. 优先 `navigator.clipboard.writeText(el.dataset.id)`；失败/不可用回退到隐藏 `textarea` + `document.execCommand('copy')`。
3. 成功后将元素文本临时改为 `已复制 ✓`，添加 `.copied` class；约 1200ms 后还原文本与 class。
4. 失败时不改变状态（保持原 ID 文本），可借助 console 提示。

使用全局函数 + 内联 `onclick`，天然兼容 htmx 无限滚动动态加载的卡片，无需监听 `htmx:afterSwap`。

### CSS：`assets/static/styles.css`

追加 `.article-id` 规则，沿用现有 `.article-time` 分隔符风格：

- `position: relative; z-index: 2;` —— 关键：`.stretched-link::after` 是 `z-index:1` 的全卡片覆盖层（用于点卡片任意处打开原文），meta 行在其下方，点击会被拦截。提升 `.article-id` 层级到 2 才能接收点击（与现有 `.article-actions` 按钮同理）。
- `::before { content: "·"; margin-right: 0.5rem; }`，与 meta 行现有分隔一致。
- `cursor: pointer`；hover 时变主题色（`var(--accent)`）。
- 等宽字体显示数字，便于辨认。
- `.article-id.copied`：短暂高亮反馈色。

### 模板重复标记：`assets/templates/partials/article-list.gohtml`

`article-list.gohtml` 内联了一份与 `article-items.gohtml` 几乎相同的文章卡片标记，用于**首页初始渲染**；`article-items.gohtml` 仅用于 htmx 无限滚动后续加载。两处 meta 行都需追加相同的 `article-id` span，否则只有滚动加载的新卡片有 ID、首屏卡片没有。

### base 模板：`assets/templates/base.gohtml`

在 `sidebar.js` 引入之后追加 `<script src="/static/article.js"></script>`。

## 改动文件

1. `assets/templates/partials/article-items.gohtml` — meta 行追加 ID span（无限滚动加载用）。
2. `assets/templates/partials/article-list.gohtml` — meta 行追加 ID span（首屏初始渲染用）。
3. `assets/static/article.js` — 新增 `copyArticleId` 函数。
4. `assets/static/styles.css` — 追加 `.article-id` 样式（含 z-index 提升）。
5. `assets/templates/base.gohtml` — 引入 `article.js`。

## 验证

- `go test ./...` 通过（无 Go 逻辑变更，回归保证）。
- `go install ./cmd/blogwatcher` 编译通过。
- Docker 重建部署后，在浏览器中：
  - 每张卡片 meta 行可见 `#<数字>`。
  - 点击 ID，文本短暂变为 `已复制 ✓`，剪贴板内容为该数字。
  - 点击 ID 不会触发打开原文链接。
  - 无限滚动加载的新卡片同样可复制。
