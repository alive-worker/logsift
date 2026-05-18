# Difficulty scoring detail (logsift)

Per spec appendix A — four dimensions 0/1/2 summed and mapped:
0–2 简单, 3–5 中等, 6–8 困难.

| # | 需求清晰度 | 修改范围 | 环境/依赖复杂度 | 验证复杂度 | 总分 | 分级 | category |
|---|:---:|:---:|:---:|:---:|:---:|:----:|----|
| 1 | 0 | 1 | 0 | 1 | 2 | 简单 | 代码生成 |
| 2 | 0 | 0 | 0 | 1 | 1 | 简单 | Bug修复/调试 |
| 3 | 0 | 0 | 0 | 0 | 0 | 简单 | 代码理解与分析 |
| 4 | 0 | 1 | 1 | 1 | 3 | 中等 | 测试 |
| 5 | 0 | 1 | 1 | 1 | 3 | 中等 | 功能迭代 |
| 6 | 0 | 1 | 1 | 1 | 3 | 中等 | 代码重构 |
| 7 | 1 | 1 | 2 | 2 | 6 | 困难 | DevOps/工程化 |
| 8 | 1 | 2 | 1 | 2 | 6 | 困难 | 功能迭代 |
| 9 | 0 | 1 | 0 | 1 | 2 | 简单 | 代码生成 |
| 10 | 0 | 0 | 0 | 0 | 0 | 简单 | 代码理解与分析 |
| 11 | 1 | 0 | 0 | 1 | 2 | 简单 | Bug修复/调试 |
| 12 | 0 | 1 | 1 | 1 | 3 | 中等 | 测试 |
| 13 | 0 | 1 | 1 | 1 | 3 | 中等 | 功能迭代 |
| 14 | 0 | 2 | 1 | 2 | 5 | 中等 | 代码重构 |
| 15 | 1 | 1 | 1 | 2 | 5 | 中等 | 功能迭代 |
| 16 | 0 | 1 | 1 | 1 | 3 | 中等 | 功能迭代 |

Distribution: 简单 6 / 中等 6 / 困难 4 (16 total) — slightly more 简单-heavy
than notebox because logsift's surface per task is smaller (most additions
touch one package).

Category counts: 代码生成 2, Bug修复/调试 2, 代码理解与分析 2, 测试 2,
代码重构 2, 功能迭代 5, DevOps/工程化 1 — all seven categories covered.

## Combined coverage across the public + private pair

| category          | notebox 出现     | logsift 出现     | 合计 |
|-------------------|------------------|------------------|:---:|
| 代码生成          | #9, #16          | #1, #9           |  4  |
| Bug 修复 / 调试   | #2               | #2, #11          |  3  |
| 代码重构          | #6, #14          | #6, #14          |  4  |
| 功能迭代          | #1, #5, #8, #13  | #5, #8, #13, #15, #16 |  9  |
| 测试              | #4, #10, #11     | #4, #12          |  5  |
| 代码理解与分析    | #3, #12          | #3, #10          |  4  |
| DevOps / 工程化   | #7, #15          | #7               |  3  |

Every category has at least 3 prompts across the pair; no category is
single-source.
