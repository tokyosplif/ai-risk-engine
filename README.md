# AI Risk Engine

A specialized microservice for intelligent transaction scoring powered by Large Language Models (LLM). It provides deep contextual analysis to identify sophisticated fraud schemes.

## Key Features
- **Dynamic Prompt Management:** Analysis logic and rules are externalized in `prompts.json`. This allows updating business rules (limits, triggers) instantly without rebuilding or restarting the service.
- **LLM Integration:** Utilizes the `llama-3.3-70b-versatile` model with deterministic settings (Temperature 0.1) for reliable financial verdicts.
- **Heuristic Guardrails:** A built-in heuristic layer that validates AI responses, ensuring mandatory blocking when critical anomalies are detected in the textual verdict.
- **Production Logging:** Structured JSON logging (`slog`) for seamless integration with modern monitoring stacks (ELK, Grafana Loki).

## AI Logic (Antifraud Engine)
The system is configured to automatically recognize and block:
- Anomalous transaction amounts exceeding $10,000.
- High-risk crypto exchange operations (Binance, Coinbase, etc.) starting from $2,000.
- Geographic anomalies and user profile mismatches.
