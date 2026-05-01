# Founder Engineer Example

This is fake data for testing and demonstration. It shows scalar-safe structured facts, role targeting, and bullet provenance fields.

Try:

```bash
go run ../../cmd/cvx normalize --input cv.yaml --variant variants/yc-founder-engineer.yaml --output expected/current.json
go run ../../cmd/cvx render --input cv.yaml --variant variants/yc-founder-engineer.yaml --format html --output expected/cv.html
```
