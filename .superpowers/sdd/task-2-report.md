# Task 2 Report: 前端骨架与 API 客户端

## 状态

已完成。

## 实现内容

- 初始化 Vite React 前端应用：`apps/web`
- 添加 Tailwind/PostCSS 配置：
  - `apps/web/tailwind.config.js`
  - `apps/web/postcss.config.js`
- 添加 API 客户端：`apps/web/src/lib/api.ts`
- 添加健康检查 UI：`apps/web/src/App.tsx`
- 添加 Vitest 测试配置与测试初始化：
  - `apps/web/vitest.config.ts`
  - `apps/web/src/setupTests.ts`
- 添加测试覆盖：
  - `apps/web/src/lib/api.test.ts`
  - `apps/web/src/App.test.tsx`

## TDD 记录

1. 先添加 `apiGet` 行为测试。
2. 运行 `npm test -- src/lib/api.test.ts`，确认失败：
   - `Failed to resolve import "./api"`
3. 补齐 `ApiEnvelope` 和 `apiGet` 实现。
4. 再次运行同一测试，API 测试通过。
5. 先添加健康检查 UI 测试。
6. 运行 `npm test -- src/App.test.tsx`，确认失败：
   - 找不到 `Hangout`
   - 找不到 `API status: unavailable`
7. 按 brief 替换 `App.tsx` 并接入 `/api/health`。
8. 再次运行同一测试，App 测试通过。

## 验证

```bash
cd apps/web
npm test
npm run build
npm run lint
```

结果：全部通过。

- `npm test`：2 个测试文件，5 个测试通过。
- `npm run build`：TypeScript 与 Vite 生产构建通过。
- `npm run lint`：无 ESLint 错误。

## 备注

- `npx tailwindcss init -p` 在 `tailwindcss@4.3.1` 下失败，因为该包不再提供 CLI `bin`；已固定为 `tailwindcss@3.4.17`，与 brief 指定的初始化命令兼容。
- `vite.config.ts` 保持只包含 Vite 构建配置；Vitest 配置拆到 `vitest.config.ts`，避免构建时混用 Vitest 内置 Vite 类型导致 TypeScript 错误。
- 安装依赖时 npm 报告 1 个 low severity vulnerability；未执行 `npm audit fix`，避免引入未要求的依赖升级。
- 仓库中已有未跟踪 `supabase/`，本任务未修改也不会纳入本次提交。
