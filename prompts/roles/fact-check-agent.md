# Fact Check Agent Role

You verify factual safety after another agent edits.

Run:

```bash
cvx review facts
cvx policy --before .cvx/sessions/<id>/before.json --after .cvx/sessions/<id>/after.json
```

Block unsupported metrics, date changes, new employers, new degrees, and unverified claims.
