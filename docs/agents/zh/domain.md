# 领域文档

工程类技能在探索本仓库代码时，应该如何读取和理解本项目的领域文档。

## 探索前必读

- 根目录下的 **`CONTEXT.md`**，或者
- 根目录下的 **`CONTEXT-MAP.md`**（如果存在）——它会指向各个子模块独立的 `CONTEXT.md` 文件。请阅读所有与当前主题相关的文档。
- **`docs/adr/`** ——阅读与您即将进行的工作区域相关的 ADR（架构决策记录）。在多上下文仓库中，还需要检查 `src/<context>/docs/adr/` 目录以获取属于该子模块的局部决策。

如果上述任何文件不存在，**请保持静默，照常继续**。不要提示或警告这些文件缺失，也不要在最开始就建议主动创建它们。生成此类文档的技能（`/grill-with-docs`）会在词汇术语或决策真正解决时，以懒加载的方式创建它们。

## 文件结构

单上下文仓库（绝大多数项目）：

```
/
├── CONTEXT.md
├── docs/adr/
│   ├── 0001-event-sourced-orders.md
│   └── 0002-postgres-for-write-model.md
└── src/
```

多上下文仓库（根目录下存在 `CONTEXT-MAP.md`）：

```
/
├── CONTEXT-MAP.md
├── docs/adr/                          ← 系统全局架构决策
└── src/
    ├── ordering/
    │   ├── CONTEXT.md
    │   └── docs/adr/                  ← 针对 ordering 子模块的局部架构决策
    └── billing/
        ├── CONTEXT.md
        └── docs/adr/
```

## 使用术语表的词汇

当您的输出涉及领域概念（例如在 Issue 标题、重构提议、排查假设、测试名中）时，请务必使用在 `CONTEXT.md` 中定义的术语。不要漂移到术语表中已被明确禁止或规避的同义词。

如果您需要的概念尚未列入术语表，这是一个信号——要么您正在发明该项目原本不使用的词汇（请重新考虑），要么确实存在表达空白（请在调用 `/grill-with-docs` 时将其记录下来）。

## 标记与 ADR 的冲突

如果您的实现或输出与已有的 ADR 冲突，请显式提及，而不要静默覆盖：

> _Contradicts ADR-0007 (event-sourced orders) — but worth reopening because…（与 ADR-0007“事件溯源订单”冲突——但值得重新讨论，因为……）_
