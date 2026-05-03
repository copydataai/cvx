# Install cvx

From a checked-out repository:

```bash
go install ./cmd/cvx
```

From a published module path, after the repository is tagged:

```bash
go install github.com/josesanchez/cvx/cmd/cvx@latest
```

Verify:

```bash
cvx version
cvx check --variant variants/yc-founder-engineer.yaml
```

`cvx` has no SaaS dependency and does not require API keys for core functionality.
