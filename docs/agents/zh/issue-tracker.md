# 问题追踪器：GitHub

本仓库的问题（Issues）和 PRD 均作为 GitHub Issues 存放。所有操作使用 `gh` 命令行工具（CLI）进行。

## 约定

- **创建问题**：`gh issue create --title "..." --body "..."`。使用 heredoc 编写多行正文。
- **读取问题**：`gh issue view <number> --comments`，配合 `jq` 过滤评论并获取标签。
- **列出问题**：`gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'`，可使用 `--label` 和 `--state` 过滤。
- **评论问题**：`gh issue comment <number> --body "..."`
- **应用/移除标签**：`gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **关闭问题**：`gh issue close <number> --comment "..."`

自动从 `git remote -v` 推断 GitHub 仓库——在克隆仓库中运行时，`gh` 会自动处理。

## 当技能提示“发布到问题追踪器”时

创建一个 GitHub Issue。

## 当技能提示“获取相关单据”时

运行 `gh issue view <number> --comments`。
