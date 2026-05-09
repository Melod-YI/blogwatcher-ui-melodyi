# Phase 19: Blog Settings Enhancement - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-09
**Phase:** 19-Blog Settings Enhancement
**Areas discussed:** URL 显示位置, 编辑交互模式, 验证反馈方式

---

## URL 显示位置

| Option | Description | Selected |
|--------|-------------|----------|
| A: 两列并排 | Blog URL 和 Feed URL 左右并列显示，信息一目了然 | ✓ |
| B: 堆叠显示 | Blog URL 在上，Feed URL 在下，垂直布局适配窄卡片 | |
| C: 点击展开 | 默认只显示 Blog URL，点击展开查看 Feed URL | |

**User's choice:** A: 两列并排
**Notes:** 用户偏好信息一目了然，左右并列对比清晰

---

## 编辑交互模式

| Option | Description | Selected |
|--------|-------------|----------|
| A: 统一编辑表单 | 点击 Edit 进入表单，统一编辑 name + URL + Feed URL + category | ✓ |
| B: 分别点击编辑 | 每个字段旁有编辑图标，单独点击进入编辑 | |
| C: URL 行直接编辑 | 点击 URL 链接直接转为输入框编辑 | |

**User's choice:** A: 统一编辑表单
**Notes:** 与现有 Edit 流程一致，用户习惯已建立，一次编辑所有字段

---

## 验证反馈方式

| Option | Description | Selected |
|--------|-------------|----------|
| A: Inline 错误提示 | 输入框下方显示红色错误文字，阻止提交 | ✓ |
| B: Toast 提示 | 顶部弹出 toast，不改变表单布局 | |
| C: 提交后验证 | 后端验证失败后刷新表单，显示错误 | |

**User's choice:** A: Inline 错误提示
**Notes:** 错误位置精确，用户立即知道哪个字段有问题，阻止提交无效数据

---

## Claude's Discretion

- URL 输入框宽度、错误提示字体大小等样式细节 — 按现有表单风格自由设计
- 后端验证错误消息具体措辞 — 与现有风格一致即可

---

## Deferred Ideas

None — discussion stayed within phase scope.