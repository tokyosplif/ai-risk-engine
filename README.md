# 🧠 AI Risk Engine

The AI Risk Engine is a specialized microservice designed for intelligent transaction scoring and contextual fraud analysis. By combining the reasoning capabilities of Large Language Models (LLMs) with a high-performance, deterministic Go heuristic layer, it provides sophisticated protection against complex fraud schemes.

## 🌟 Key Features
* **Hybrid Analysis Pipeline:** Implements a dual-layer decision engine that correlates LLM reasoning with hard-coded heuristic safety checks (Business Rules Abstraction).
* **Deterministic AI Verdicts:** Utilizes the `llama-3.3-70b-versatile` model with a temperature of `0.1` and forced JSON-mode output for consistent, reliable financial assessments.
* **Idempotent Tagging:** Automatically normalizes AI outputs, ensuring alert tags (e.g., `[PENDING REVIEW]`) are applied cleanly without duplication.
* **Hot-Reloadable Prompts:** Uses `fsnotify` to watch `prompts.json` for changes. Business rules, security protocols, and few-shot examples can be updated instantly without service restarts or recompilation.

## ⚙️ Architecture & Logic

### The Analyzer (Heuristic Layer)
The engine applies a multi-stage validation process to every AI verdict. All magic numbers are extracted into business-rule constants for easy tuning:
1. **Low Value Pass (`< $500`):** Automatically overrides AI blocks for minor transactions, preventing false positives for everyday purchases.
2. **Strict Anomaly Blocking:** Forces a `[Heuristic Block]` if a transaction exceeds $500 AND is more than 2x the user's historical maximum, protecting compromised accounts.
3. **Confidence Mapping:**
   * **High Confidence (>75%):** Trusts the AI's verdict entirely.
   * **Medium Confidence (31–75%):** Downgrades "Block" verdicts on massive amounts (`> $10,000`) to `[PENDING REVIEW]` to prevent aggressive false positives while alerting human operators.
   * **Low Confidence (≤30%):** Automatically treats the verdict as non-blocking (`[Low Confidence Ignore]`).

### Prompt Management (`prompts.json`)
The system's "intelligence" is externalized into a dynamic configuration:
* **Historical Context:** Instructs the AI to normalize behavior for users with zero-stats (Cold Start).
* **Crypto/P2P Policy:** Specific protocols for high-risk merchants (e.g., Binance, Coinbase), requiring both high amounts and geographic anomalies for a block.
* **Few-Shot Examples:** Includes training pairs in the prompt to ensure the AI understands the difference between a "Safe Cold Start" and a "High-Value Mismatch."

## 🛠️ Technical Specifications
* **Communication:** gRPC for low-latency inter-service calls with built-in retries and timeouts.
* **Language:** Go 1.25.
* **AI Provider:** Groq / OpenAI compatible API.
* **Testing:** Fully testable architecture using Mock LLM clients to validate heuristic edge cases without hitting external APIs.
* **Logging:** Structured JSON logging (`slog`) with dynamic log-level configuration.
