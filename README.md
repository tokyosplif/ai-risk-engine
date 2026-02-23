# AI Risk Engine

The AI Risk Engine is a specialized microservice designed for intelligent transaction scoring and contextual fraud analysis. By combining the reasoning capabilities of Large Language Models (LLMs) with a high-performance heuristic safety layer, it provides sophisticated protection against complex fraud schemes.

## Key Features
- **Hot-Reloadable Prompts:** Uses fsnotify to watch prompts.json for changes. Business rules, security protocols, and few-shot examples can be updated instantly without service restarts or recompilation.
- **Hybrid Analysis Pipeline:** Implements a dual-layer decision engine that correlates LLM reasoning with hard-coded heuristic safety checks.
- **Deterministic AI Verdicts:** Utilizes the `llama-3.3-70b-versatile` model with a temperature of 0.1 and JSON-mode forced output for consistent, reliable financial assessments.
- **Cold Start Adaptation:** Intelligently handles new users with zero transaction history, preventing false positives for low-value local purchases.

## Architecture & Logic
1. **The Analyzer (Heuristic Layer)**
The engine applies a multi-stage validation process to every AI verdict:

- **Cold Start Protection:** Automatically overrides AI blocks for transactions under $500 if the user has no history (MaxTx: 0) and the merchant is local.
- **Strict Anomaly Blocking:** Forces a block if a transaction exceeds $500 AND is more than 2x the user's historical maximum.
- **Confidence Mapping:**
  - High Confidence (>85%): Grants "VIP Trust" to high-value transactions approved by AI.
  - Medium Confidence (31–75%): Downgrades "Block" verdicts to [PENDING REVIEW] to prevent aggressive false positives.
  - Low Confidence (≤30%): Automatically treats the verdict as non-blocking.

2. **Prompt Management (prompts.json)**
The system's "intelligence" is externalized into a dynamic configuration:

- **Historical Context:** Instructs the AI to normalize behavior for users with zero-stats.
- **Crypto/P2P Policy:** Specific protocols for high-risk merchants (e.g., Binance, Coinbase), requiring both high amounts (>$1000) and geographic anomalies for a block.
- **Few-Shot Examples:** Includes training pairs in the prompt to ensure the AI understands the difference between a "Safe Cold Start" and a "Lagos High-Value Mismatch."

## Security Protocols (Antifraud Engine)
The AI is explicitly instructed to follow these prioritized rules:

- **Normalization:** Transactions matching local geography (e.g., Lviv, Kyiv) and standard merchants (e.g., Supermarkets) are marked as NORMAL even if they slightly exceed averages.
- **High-Value Buffer:** Transactions > $10,000 with consistent locations are flagged for PENDING REVIEW rather than immediate blocking.
- **Critical Mismatch:** Only triggers is_blocked: true when there is a 75%+ fraud probability, such as a geographic mismatch combined with an unknown offshore merchant.

## Technical Specifications
- **Communication:** gRPC for low-latency inter-service calls.
- **Language:** Go 1.25.
- **AI Provider:** Groq/OpenAI compatible API.
- **Logging:** Structured JSON (slog) including pattern-matching alerts for heuristic triggers (e.g., "binance", "nigeria", "anomaly").
