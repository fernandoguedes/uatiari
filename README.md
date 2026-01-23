# uatiari 🎯

> **uatiari** (Nheengatu: *to guide*) — An intelligent code review agent that guides developers toward better code quality through eXtreme Programming principles.

---

## ✨ Overview

**uatiari** analyzes your git branches using AI and provides structured feedback based on XP best practices:

- 🧪 **Test-Driven Development** — Ensures critical paths have test coverage
- 🎨 **Simple Design** — Identifies unnecessary complexity and abstractions
- 🔍 **Code Smells** — Detects duplication, god classes, and deep nesting
- 🚫 **YAGNI** — Flags premature optimization and speculative code
- ⚡ **Business Logic** — Validates domain rules and edge cases

### The Philosophy

Just as a guide leads travelers through complex terrain, **uatiari** guides developers through code reviews with:

1. **Plan** → The agent analyzes your diff and proposes a review strategy
2. **Approval** → You explicitly approve before execution (human-in-the-loop)
3. **Execution** → Receives structured, actionable feedback

---

## 🚀 Quick Start

### Prerequisites

```bash
Python 3.11+  |  Git  |  Google Gemini API Key
```

### Installation

#### Standard Installation (Recommended)

Install the standalone binary (no Python required) using our installer script:

```bash
curl -fsSL https://raw.githubusercontent.com/fernandoguedes/uatiari/main/install.sh | bash
```

This will install `uatiari` to `~/.local/bin`.

#### Development Installation

```bash
# Clone and setup
git clone https://github.com/fernandoguedes/uatiari.git
cd uatiari
poetry install
```

### Configuration

`uatiari` looks for your `GOOGLE_API_KEY` in the following locations (highest priority first):

1. **Local .env**: `./.env` (Project specific overrides)
2. **Global Config**: `~/.config/uatiari/.env` (Recommended for global use)
3. **Legacy Config**: `~/.uatiari.env`
4. **Environment Variable**: `GOOGLE_API_KEY` exported in shell

**Setup Global Configuration:**

```bash
mkdir -p ~/.config/uatiari
echo "GOOGLE_API_KEY=your-key-here" > ~/.config/uatiari/.env
```

> 🔑 Get your API key at [Google AI Studio](https://aistudio.google.com/app/apikey)

### Basic Usage

```bash
# Review a feature branch
uatiari feature/user-authentication

# Compare against a different base
uatiari feature/new-api --base=develop
```

### Updating

Update to the latest version directly from the CLI:

```bash
uatiari update
```

---

## 📊 Example Session

```bash
$ uatiari feature/payment-validation
```

```
🎯 uatiari - XP Code Reviewer
📊 Branch: feature/payment-validation (base: main)

⏳ Fetching git context...
✅ Found 3 changed file(s)

⏳ Generating review plan...
```

<details>
<summary>📋 Review Plan</summary>

```
╭─────────────────────────────────────────────────────────────────╮
│                                                                 │
│  🔴 HIGH RISK                                                    │
│  • src/payment/processor.py (145L) - payment validation         │
│                                                                 │
│  🟡 MEDIUM RISK                                                  │
│  • src/services/email.py (34L) - notification handling          │
│                                                                 │
│  🟢 LOW RISK                                                     │
│  • tests/test_payment.py (89L) - test coverage                  │
│                                                                 │
│  XP Focus:                                                      │
│  • Business rule correctness in amount validation               │
│  • Simple design: single responsibility check                   │
│  • Test coverage for critical payment paths                     │
│                                                                 │
│  Estimated time: 8-10 minutes                                   │
│                                                                 │
╰─────────────────────────────────────────────────────────────────╯
```

</details>

```
Approve execution? (y/n): y

🚀 Executing XP review...
```

<details>
<summary>✅ Review Results</summary>

```json
{
  "blocking_issues": [
    {
      "file": "src/payment/processor.py",
      "lines": "78-92",
      "category": "BUSINESS_LOGIC",
      "issue": "Payment validation allows negative amounts",
      "action": "Add validation: amount > 0 and < MAX_TRANSACTION_LIMIT",
      "why_blocking": "Risk of fraudulent transactions"
    }
  ],
  "warnings": [
    {
      "file": "src/payment/processor.py",
      "lines": "120-165",
      "category": "COMPLEXITY",
      "issue": "45-line method handles validation + API call + persistence",
      "suggestion": "Extract 'persistTransaction' to separate method",
      "effort": "20 minutes",
      "xp_principle": "Simple Design"
    }
  ],
  "test_analysis": {
    "has_tests": true,
    "test_files": ["tests/test_payment.py"],
    "notes": "Missing edge case: zero and negative amount tests",
    "verdict": "NEEDS_IMPROVEMENT"
  },
  "overall": {
    "verdict": "REQUEST_CHANGES",
    "reason": "Critical business validation missing",
    "confidence": "HIGH"
  }
}
```

</details>

---

## 🏗️ Architecture

### Workflow

```mermaid
graph LR
    A[START] --> B[Fetch Git Context]
    B --> C[Generate Plan]
    C --> D{Human Approval?}
    D -->|Yes| E[Execute Review]
    D -->|No| F[END]
    E --> G[Generate Report]
    G --> F
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| **Orchestration** | LangGraph (state machine) |
| **AI Model** | Google Gemini 2.0 Flash |
| **Git Integration** | Native Git CLI |
| **Terminal UI** | Rich library |
| **Language** | Python 3.11+ |

### Project Structure

```
uatiari/
├── 📁 src/
│   ├── cli.py                 # Entry point
│   ├── config.py              # Environment setup
│   ├── 📁 graph/
│   │   ├── state.py           # Workflow state
│   │   ├── nodes.py           # LangGraph nodes
│   │   └── workflow.py        # State machine
│   ├── 📁 tools/
│   │   └── git_tools.py       # Git operations
│   └── 📁 prompts/
│       └── xp_reviewer.py     # System prompts
└── 📁 tests/                  # Test suite
```

---

## 📖 Output Format

### Blocking Issues

Critical problems that **must** be fixed:

- ❌ Business logic violations
- 🔒 Security vulnerabilities  
- 💾 Data corruption risks

### Warnings

Important issues that **should** be addressed:

- ⚠️ Code complexity (god methods, deep nesting)
- 🧪 Missing tests for critical paths
- 📋 Code duplication

### Suggestions

Optional improvements:

- ✨ Naming clarity
- 🔧 Small refactorings (<30min)
- 🚫 YAGNI violations

### Verdicts

| Verdict | Meaning |
|---------|---------|
| ✅ **APPROVE** | Ready to merge |
| 🔄 **REQUEST_CHANGES** | Needs fixes before merge |
| 🛑 **BLOCK** | Critical issues present |

---

## 🛠️ Development

### Run Tests

```bash
poetry run pytest -v
```

### Code Quality

```bash
# Format code
poetry run black src/ tests/

# Lint
poetry run ruff check src/ tests/
```

### Customize XP Rules

Edit `src/prompts/xp_reviewer.py` to modify:
- Review priorities
- Blocking conditions
- XP principles enforced

---

## 🐛 Troubleshooting

<details>
<summary><strong>Common Issues</strong></summary>

| Problem | Solution |
|---------|----------|
| `Not in a git repository` | Run from within a git repo directory |
| `Branch does not exist` | Verify with `git branch -a` |
| `GOOGLE_API_KEY not found` | Add key to `.env` file |
| `No differences found` | Branches are identical |
| Review seems incomplete | Large diffs may be truncated |

</details>

---

## 🎓 XP Principles

### What We Enforce

| Principle | Implementation | Action |
|-----------|---------------|--------|
| **Test-Driven Development** | Production code needs tests | 🛑 BLOCK if missing |
| **Simple Design** | No god classes, deep nesting | ⚠️ WARN if complex |
| **Refactoring** | Small, safe improvements | 💡 SUGGEST steps |
| **YAGNI** | No premature optimization | 🚫 FLAG violations |

---

## 🤝 Contributing

This project follows XP values:

- ✅ **Tests first** — TDD approach
- 🎯 **Simplicity** — YAGNI, Simple Design
- 🔄 **Continuous refactoring** — Small improvements
- 📦 **Small commits** — Focused changes

**PR Guidelines:**
1. Include tests demonstrating the change
2. Keep implementation simple and focused
3. Write clear commit messages (explain WHY)

---

## 📄 License

MIT License — see [LICENSE](LICENSE) file for details

---

## 🙏 Acknowledgments

Built with:
- [LangGraph](https://github.com/langchain-ai/langgraph) — State machine orchestration
- [Google Gemini](https://ai.google.dev/) — AI-powered code analysis
- [Rich](https://github.com/Textualize/rich) — Beautiful terminal output

---

<div align="center">

**"A guide does not carry you — they show you the path."**

*Made with ❤️ by developers, for developers*

</div>
