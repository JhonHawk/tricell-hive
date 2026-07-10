
## Python Standards

- **Check the project's Python version** (`pyproject.toml` → `requires-python`, `.python-version`, or Dockerfile) before using modern syntax. `str | None` requires 3.10+. `match/case` requires 3.10+. `type` statement requires 3.12+. Use `from __future__ import annotations` (available 3.7+) to defer annotation evaluation on 3.9.
- **Prefer `str | None` over `Optional[str]`** when the project targets 3.10+. Use builtin generics: `list[str]` not `List[str]`, `dict[str, int]` not `Dict[str, int]`.
- **PEP 695 `type` statement (3.12+)**: Prefer `type Point = tuple[float, float]` over `TypeAlias`. Use `class Stack[T]:` over `class Stack(Generic[T])`.
- **Ruff is the linter AND formatter** — never generate separate flake8, isort, or black configs. But ruff is NOT a type checker: pair with `mypy` or `Pyright`. Run both: `ruff check && mypy .`
- **Project metadata lives in `pyproject.toml`**, not `setup.py` or `requirements.txt`. Generate `requirements.txt` only as a deployment artifact (Lambda layers, Docker images, GH Actions matrices) using `uv pip compile pyproject.toml -o requirements.txt`. Never edit the generated file by hand; never use it as the source of truth.
- **Optional-arg reset uses `is None`, not `or`:** `if items is None: items = []` — `items = items or []` silently discards an explicitly-passed empty list or other falsy value.
- **Dataclasses vs Pydantic**: `dataclasses` for internal data structures, `pydantic.BaseModel` for external data (API requests/responses, config, env validation). Never use plain dicts for structured data crossing function boundaries.
- **pytest fixtures**: use `yield` for resource cleanup, factory pattern for multi-instance tests, share fixtures via `conftest.py`, justify scope (`session`/`module`) for expensive resources.
- **Async**: prefer `asyncio` with `async/await` over threading for I/O-bound work. Use `asyncio.TaskGroup` (3.11+) over `gather()` for structured concurrency.
