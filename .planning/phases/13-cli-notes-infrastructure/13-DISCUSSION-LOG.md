# Phase 13: CLI Notes Infrastructure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-07
**Phase:** 13-CLI Notes Infrastructure
**Areas discussed:** 命令结构, 文章验证, 覆盖行为, 错误反馈, 目录创建, 删除不存在, 文件处理

---

## 命令结构

| Option | Description | Selected |
|--------|-------------|----------|
| 顶层命令 note (推荐) | ROADMAP 明确指定 'note --article-id'，保持简单直接的命令风格 | ✓ |
| article 子命令 | 与现有 article list/mark-read/mark-unread 命令保持一致的组织结构 | |

**User's choice:** 顶层命令 note (推荐)
**Notes:** None

---

## 文章验证

| Option | Description | Selected |
|--------|-------------|----------|
| 必须验证存在 (推荐) | 确保备注与实际文章关联，避免无效 ID。但需要数据库查询开销。 | ✓ |
| 不验证，直接写入 | 允许为任意 ID 写入备注，更灵活但可能导致孤立备注文件。 | |

**User's choice:** 必须验证存在 (推荐)
**Notes:** None

---

## 覆盖行为

| Option | Description | Selected |
|--------|-------------|----------|
| 静默覆盖 (推荐) | 简单直接，符合备注可能需要更新的使用场景。 | ✓ |
| 提示确认覆盖 | 提醒用户已有备注，让用户确认是否覆盖。增加交互步骤。 | |

**User's choice:** 静默覆盖 (推荐)
**Notes:** None

---

## 错误反馈

| Option | Description | Selected |
|--------|-------------|----------|
| 简单报错 (推荐) | 简单错误消息 + 退出，与现有 CLI 命令风格一致。 | ✓ |
| 详细用法提示 | 详细说明缺少哪些参数，并显示命令用法示例。更友好但增加代码量。 | |

**User's choice:** 简单报错 (推荐)
**Notes:** None

---

## 目录创建

| Option | Description | Selected |
|--------|-------------|----------|
| 自动创建目录 (推荐) | 用户体验好，第一次写入备注时自动创建目录，无需手动准备 | ✓ |
| 报错退出 | 更安全，需要用户预先创建目录，避免意外创建 | |

**User's choice:** 自动创建目录 (推荐)
**Notes:** None

---

## 删除不存在备注

| Option | Description | Selected |
|--------|-------------|----------|
| 报错提示 (推荐) | 一致的用户体验，告知用户操作无效，避免误解 | ✓ |
| 静默成功 | 简单处理，如果文件不存在则操作成功（无副作用） | |

**User's choice:** 报错提示 (推荐)
**Notes:** None

---

## 文件处理方式

| Option | Description | Selected |
|--------|-------------|----------|
| 仅支持 UTF-8 (推荐) | 简单实现，符合 Markdown 文件常见编码约定 | |
| 支持多种编码 | 更全面，但增加编码检测和转换的复杂度 | |

**User's choice:** 希望直接复制文件，而非读写处理
**Notes:** 用户希望使用字节级文件复制（io.Copy 或 os.ReadFile + os.WriteFile），避免编码转换问题。此方案更简单高效。

---

## Claude's Discretion

None — 用户对所有决策提供了明确选择。

## Deferred Ideas

None — 讨论始终保持在阶段范围内。